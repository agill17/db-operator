package aws

import (
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
)

func (R RDSClient) CreateDBCluster(in *v1alpha1.DBCluster) error {
	if _, errCreating := R.rdsClient.CreateDBCluster(inputToCreateDBClusterInput(in)); errCreating != nil {
		return errCreating
	}
	return nil
}

func (R RDSClient) DeleteDBCluster(in *v1alpha1.DBCluster) error {
	if _, errDeleting := R.rdsClient.DeleteDBCluster(inputToDeleteDBClusterInput(in)); errDeleting != nil {
		if awsErr, isAwsErr := errDeleting.(awserr.Error); isAwsErr {
			if awsErr.Error() == rds.ErrCodeDBClusterNotFoundFault { // if for some reason the dbCluster is not found, ignore and move on
				return nil
			}
		}
		return errDeleting
	}
	return nil
}

func (R RDSClient) ModifyDBCluster(in *v1alpha1.DBCluster) error {
	if _, errUpdating := R.rdsClient.ModifyDBCluster(inputToModifyDBClusterInput(in)); errUpdating != nil {
		return errUpdating
	}
	return nil
}

func (R RDSClient) DBClusterExists(in *v1alpha1.DBCluster) (bool, error) {
	_, err := R.rdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(inputToDBClusterID(in)),
	})
	if err != nil {
		if awsErr, isAwsErr := err.(awserr.Error); isAwsErr {
			if awsErr.Error() == rds.ErrCodeDBClusterNotFoundFault {
				return false, nil
			}
		}
		return false, err
	}
	return true, nil
}

func (R RDSClient) IsDBClusterUpToDate(in *v1alpha1.DBCluster) (bool, error) {
	panic("implement me")
}

func inputToCreateDBClusterInput(in *v1alpha1.DBCluster) *rds.CreateDBClusterInput {
	return nil
}
func inputToDeleteDBClusterInput(in *v1alpha1.DBCluster) *rds.DeleteDBClusterInput {
	return nil
}
func inputToModifyDBClusterInput(in *v1alpha1.DBCluster) *rds.ModifyDBClusterInput {
	return nil
}

func inputToDBClusterID(in *v1alpha1.DBCluster) string {
	clusterID := fmt.Sprintf("%s-%s", in.GetNamespace(), in.GetName())
	if in.Spec.DBClusterIdentifierOverride != "" {
		clusterID = in.Spec.DBClusterIdentifierOverride
	}
	return clusterID
}
