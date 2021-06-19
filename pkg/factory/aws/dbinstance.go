package aws

import (
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
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

func (i InternalAwsClients) ModifyDBInstance(input *v1alpha1.DBInstance, password string) error {
	panic("implement me")
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
		if resp.DBInstances[0].Endpoint.Address != nil && *resp.DBInstances[0].Endpoint.Address != "" {
			out.Endpoint = *resp.DBInstances[0].Endpoint.Address
		}
	}
	return out, nil
}

func (i InternalAwsClients) IsDBInstanceUpToDate(input *v1alpha1.DBInstance) (bool, error) {
	panic("implement me")
}
