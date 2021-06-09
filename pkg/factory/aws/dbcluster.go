package aws

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/google/go-cmp/cmp"
)

const (
	// the why - because I cannot find rds.ErrCodeInvalidParameterCombination :/
	deletionProtectionErrMessage = "Cannot delete protected Cluster, please disable deletion protection and try again."
)

func (i InternalAwsClients) CreateDBCluster(input *v1alpha1.DBCluster, password string) error {
	_, errCreating := i.rdsClient.CreateDBCluster(createDBClusterInput(input, password))
	return errCreating
}

func (i InternalAwsClients) DeleteDBCluster(dbCluster *v1alpha1.DBCluster) error {
	namespacedName := fmt.Sprintf("%s/%s", dbCluster.GetNamespace(), dbCluster.GetName())
	// dont even make a delete attempt if deletionProtection is enabled in CR spec.
	if dbCluster.Spec.DeletionProtection {
		return ErrDBClusterDeletionProtectionEnabled{Message: fmt.Sprintf(
			"%v/%v: Cannot delete, deletion protection is enabled", dbCluster.GetNamespace(),
			dbCluster.GetName())}
	}

	if _, errDeleting := i.rdsClient.DeleteDBCluster(deleteDBClusterInput(dbCluster)); errDeleting != nil {
		if awsErr, isAwsErr := errDeleting.(awserr.Error); isAwsErr {
			switch awsErr.Error() {
			case rds.ErrCodeDBClusterNotFoundFault:
				i.logger.Info(fmt.Sprintf("%v - does not exist, nothing to delete.", namespacedName))
				return nil
			}
			// if the error message says the following,
			// attempt to do a update in case user changed the deletionProtection after deleting CR
			if awsErr.Message() == deletionProtectionErrMessage {
				i.logger.Info(fmt.Sprintf("%v - has deletionProtection enabled in AWS, checking if updating can resolve this.", namespacedName))
				isUpToDate, modifyIn, err := i.IsDBClusterUpToDate(dbCluster)
				if err != nil {
					return err
				}
				if !isUpToDate {
					if errUpdating := i.ModifyDBCluster(modifyIn); errUpdating != nil {
						return errUpdating
					}
					// TODO: catch this error in dbcluster_controller and quietly requeue
					return ErrRequeueNeeded{Message: fmt.Sprintf("ErrRequeueNeededToRetryDeleteAfterUpdate")}
				}
			}
		}
		return errDeleting
	}
	return nil
}

func (i InternalAwsClients) ModifyDBCluster(modifyIn interface{}) error {
	rdsModifyIn, ok := modifyIn.(*rds.ModifyDBClusterInput)
	if !ok {
		errMsg := fmt.Sprintf("Expected Type: %T but got %T in AWS ModifyDBCluster implementation", &rds.ModifyDBClusterInput{}, modifyIn)
		return ErrInvalidTypeWasPassedIn{Message: errMsg}
	}
	rdsModifyIn.ApplyImmediately = aws.Bool(true)
	if _, errUpdating := i.rdsClient.ModifyDBCluster(rdsModifyIn); errUpdating != nil {
		return errUpdating
	}
	return nil
}

func (i InternalAwsClients) DBClusterExists(dbClusterID string) (bool, string, error) {
	out, err := i.rdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(dbClusterID),
	})
	if err != nil {
		if awsErr, isAwsErr := err.(awserr.Error); isAwsErr {
			if awsErr.Code() == rds.ErrCodeDBClusterNotFoundFault {
				return false, "", nil
			}
		}
		return false, "", err
	}
	return true, *out.DBClusters[0].Status, nil
}

// TODO: refactor, I am not proud of this..
func (i InternalAwsClients) IsDBClusterUpToDate(input *v1alpha1.DBCluster) (bool, interface{}, error) {
	clusterState, err := i.rdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(input.GetDBClusterID())})
	if err != nil {
		return false, nil, err
	}
	if len(clusterState.DBClusters) != 1 {
		return false, nil, errors.New("ErrMultipleDBClustersExistsWithTheSameID")
	}
	currentState := clusterState.DBClusters[0]
	modifyDBClusterInput := &rds.ModifyDBClusterInput{}
	if *currentState.EngineVersion != input.Spec.EngineVersion {
		modifyDBClusterInput.EngineVersion = aws.String(input.Spec.EngineVersion)
	}
	if *currentState.HttpEndpointEnabled != input.Spec.EnableHttpEndpoint {
		modifyDBClusterInput.EnableHttpEndpoint = aws.Bool(input.Spec.EnableHttpEndpoint)
	}
	if currentState.GlobalWriteForwardingRequested != nil && *currentState.GlobalWriteForwardingRequested != input.Spec.EnableGlobalWriteForwarding {
		modifyDBClusterInput.EnableGlobalWriteForwarding = aws.Bool(input.Spec.EnableGlobalWriteForwarding)
	}
	if currentState.DeletionProtection != nil && *currentState.DeletionProtection != input.Spec.DeletionProtection {
		modifyDBClusterInput.DeletionProtection = aws.Bool(input.Spec.DeletionProtection)
	}
	if currentState.CopyTagsToSnapshot != nil && *currentState.CopyTagsToSnapshot != input.Spec.CopyTagsToSnapshot {
		modifyDBClusterInput.CopyTagsToSnapshot = aws.Bool(input.Spec.CopyTagsToSnapshot)
	}
	if currentState.BackupRetentionPeriod != nil && *currentState.BackupRetentionPeriod != input.Spec.BackupRetentionPeriod {
		modifyDBClusterInput.BackupRetentionPeriod = aws.Int64(input.Spec.BackupRetentionPeriod)
	}
	if currentState.DBClusterParameterGroup != nil && *currentState.DBClusterParameterGroup != input.Spec.DBClusterParameterGroupName {
		modifyDBClusterInput.DBClusterParameterGroupName = aws.String(input.Spec.DBClusterParameterGroupName)
	}
	if *currentState.DBClusterIdentifier != input.GetDBClusterID() {
		modifyDBClusterInput.NewDBClusterIdentifier = aws.String(input.GetDBClusterID())
	}
	if currentState.Port != nil && input.Spec.Port != 0 && *currentState.Port != input.Spec.Port {
		modifyDBClusterInput.Port = aws.Int64(input.Spec.Port)
	}
	if currentState.PreferredBackupWindow != nil && *currentState.PreferredBackupWindow != input.Spec.PreferredBackupWindow {
		modifyDBClusterInput.PreferredBackupWindow = aws.String(input.Spec.PreferredBackupWindow)
	}
	if currentState.PreferredMaintenanceWindow != nil && *currentState.PreferredMaintenanceWindow != input.Spec.PreferredMaintenanceWindow {
		modifyDBClusterInput.PreferredMaintenanceWindow = aws.String(input.Spec.PreferredMaintenanceWindow)
	}
	// TODO: check VPC-SG and note that AWS by default adds a security group so beware when comparing with desired state with empty vpc-sg

	isUpToDate := cmp.Equal(modifyDBClusterInput, &rds.ModifyDBClusterInput{})
	if !isUpToDate {
		modifyDBClusterInput.DBClusterIdentifier = aws.String(input.GetDBClusterID())
	}
	return isUpToDate, modifyDBClusterInput, nil
}
