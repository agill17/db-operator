package aws

import (
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/google/go-cmp/cmp"
	"strings"
)

func (i InternalAwsClients) CreateDBInstance(input *v1alpha1.DBInstance, password string) error {
	_, err := i.rdsClient.CreateDBInstance(createInstanceInput(input, password))
	return err
}

func (i InternalAwsClients) DeleteDBInstance(input *v1alpha1.DBInstance) error {
	nsName := fmt.Sprintf("%s/%s", input.Namespace, input.Name)
	if input.Spec.DeletionProtection {
		msg := fmt.Sprintf("%s - cannot delete instance when deletion protection is enabled", nsName)
		return ErrDBInstanceDeletionProtectionEnabled{Message: msg}
	}
	if _, errDeleting := i.rdsClient.DeleteDBInstance(deleteDbInstanceInput(input)); errDeleting != nil {
		if awsErr, isAwsErrType := errDeleting.(awserr.Error); isAwsErrType {
			if awsErr.Error() == rds.ErrCodeDBInstanceNotFoundFault {
				i.logger.Info(fmt.Sprintf("%s - instance not found, skipping cleanup", nsName))
				return nil
			}
		}
		return errDeleting
	}
	return nil
}

func (i InternalAwsClients) ModifyDBInstance(modifyIn interface{}) error {
	rdsModifyDBInstanceIn, ok := modifyIn.(*rds.ModifyDBInstanceInput)
	if !ok {
		return ErrInvalidTypeWasPassedIn{Message: fmt.Sprintf("Expected rds.ModifyDBInstanceInput but got: %T", modifyIn)}
	}
	rdsModifyDBInstanceIn.ApplyImmediately = aws.Bool(true)
	_, err := i.rdsClient.ModifyDBInstance(rdsModifyDBInstanceIn)
	return err
}

func (i InternalAwsClients) DBInstanceExists(input *v1alpha1.DBInstance) (*v1alpha1.DBStatus, error) {
	resp, err := i.rdsClient.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(input.GetDBInstanceID()),
	})
	out := &v1alpha1.DBStatus{}
	if err != nil {
		if awsErr, isAwsErrType := err.(awserr.Error); isAwsErrType {
			if awsErr.Code() == rds.ErrCodeDBInstanceNotFoundFault {
				return out, nil
			}
		}
		return nil, err
	}

	if resp != nil && len(resp.DBInstances) == 1 {
		out.CurrentPhase = *resp.DBInstances[0].DBInstanceStatus
		out.Exists = true
		if resp.DBInstances[0].Endpoint != nil && *resp.DBInstances[0].Endpoint.Address != "" {
			out.Endpoint = *resp.DBInstances[0].Endpoint.Address
		}
	}
	return out, nil
}

func (i InternalAwsClients) IsDBInstanceUpToDate(input *v1alpha1.DBInstance) (bool, interface{}, error) {
	currentState, err := i.rdsClient.DescribeDBInstances(&rds.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(input.GetDBInstanceID()),
	})
	if err != nil {
		return false, nil, err
	}
	desiredSpec := input.Spec
	currentDBInstance := currentState.DBInstances[0]

	modifyIn := &rds.ModifyDBInstanceInput{}
	if desiredSpec.DBClusterID == "" {
		if *currentDBInstance.AllocatedStorage != desiredSpec.AllocatedStorage {
			// For MariaDB, MySQL, Oracle, and PostgreSQL, the value supplied must be at
			// least 10% greater than the current value. Values that are not at least 10%
			// greater than the existing value are rounded up so that they are 10% greater
			// than the current value.
			if desiredSpec.Engine == "mariadb" ||
				desiredSpec.Engine == "mysql" ||
				strings.HasPrefix(desiredSpec.Engine, "oracle") ||
				desiredSpec.Engine == "postgres" {
				minDesiredStorageNeeded := *currentDBInstance.AllocatedStorage + (int64(0.10 * float64(desiredSpec.AllocatedStorage)))
				if desiredSpec.AllocatedStorage >= minDesiredStorageNeeded {
					modifyIn.AllocatedStorage = aws.Int64(desiredSpec.AllocatedStorage)
				}
			} else {
				modifyIn.AllocatedStorage = aws.Int64(desiredSpec.AllocatedStorage)
			}
		}
		if *currentDBInstance.DeletionProtection != desiredSpec.DeletionProtection {
			modifyIn.DeletionProtection = aws.Bool(desiredSpec.DeletionProtection)
		}
	}

	if *currentDBInstance.AutoMinorVersionUpgrade != desiredSpec.AutoMinorVersionUpgrade {
		modifyIn.AutoMinorVersionUpgrade = aws.Bool(desiredSpec.AutoMinorVersionUpgrade)
	}

	// TODO: do a deeper comparison
	if len(currentDBInstance.EnabledCloudwatchLogsExports) != len(desiredSpec.CloudwatchLogsExports) {
		modifyIn.CloudwatchLogsExportConfiguration.EnableLogTypes = aws.StringSlice(desiredSpec.CloudwatchLogsExports)
	}
	if *currentDBInstance.DBInstanceClass != desiredSpec.DBInstanceClass {
		modifyIn.DBInstanceClass = aws.String(desiredSpec.DBInstanceClass)
	}

	if desiredSpec.Port != 0 && *currentDBInstance.DbInstancePort != desiredSpec.Port {
		modifyIn.DBPortNumber = aws.Int64(desiredSpec.Port)
	}
	if len(currentDBInstance.DBSecurityGroups) != len(desiredSpec.DBSecurityGroups) {
		modifyIn.DBSecurityGroups = aws.StringSlice(desiredSpec.DBSecurityGroups)
	}

	if *currentDBInstance.PerformanceInsightsEnabled != desiredSpec.EnablePerformanceInsights {
		modifyIn.EnablePerformanceInsights = aws.Bool(desiredSpec.EnablePerformanceInsights)
	}
	if *currentDBInstance.EngineVersion != desiredSpec.EngineVersion {
		modifyIn.EngineVersion = aws.String(desiredSpec.EngineVersion)
	}
	if *currentDBInstance.PubliclyAccessible != desiredSpec.PubliclyAccessible {
		modifyIn.PubliclyAccessible = aws.Bool(desiredSpec.PubliclyAccessible)
	}
	if desiredSpec.StorageType != "" && *currentDBInstance.StorageType != desiredSpec.StorageType {
		modifyIn.StorageType = aws.String(desiredSpec.StorageType)
	}

	isUpToDate := cmp.Equal(modifyIn, &rds.ModifyDBInstanceInput{})
	if !isUpToDate {
		modifyIn.DBInstanceIdentifier = aws.String(input.GetDBInstanceID())
	}
	return isUpToDate, modifyIn, nil
}
