package aws

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func clientObjToDBCluster(obj client.Object) (*v1alpha1.DBCluster, error) {
	dbCluster, ok := obj.(*v1alpha1.DBCluster)
	if !ok {
		return nil, errors.New(fmt.Sprintf("ErrCasting%TtoDBCluster", obj))
	}
	return dbCluster, nil
}

func (i InternalAwsClients) CreateDBCluster(input *v1alpha1.DBCluster, password string) error {
	_, errCreating := i.rdsClient.CreateDBCluster(createDBClusterInput(input, password))
	return errCreating
}

func (i InternalAwsClients) DeleteDBCluster(input *v1alpha1.DBCluster) error {
	dbCluster, errCasting := clientObjToDBCluster(input)
	if errCasting != nil {
		return errCasting
	}
	if _, errDeleting := i.rdsClient.DeleteDBCluster(deleteDBClusterInput(dbCluster)); errDeleting != nil {
		if awsErr, isAwsErr := errDeleting.(awserr.Error); isAwsErr {
			if awsErr.Error() == rds.ErrCodeDBClusterNotFoundFault { // if for some reason the dbCluster is not found, ignore and move on
				return nil
			}
		}
		return errDeleting
	}
	return nil
}

func (i InternalAwsClients) ModifyDBCluster(input *v1alpha1.DBCluster, password string) error {

	if _, errUpdating := i.rdsClient.ModifyDBCluster(modifyDBClusterInput(input, password)); errUpdating != nil {
		return errUpdating
	}
	return nil
}

func (i InternalAwsClients) DBClusterExists(input *v1alpha1.DBCluster) (bool, error) {

	_, err := i.rdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(input.GetDBClusterID()),
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

func (i InternalAwsClients) IsDBClusterUpToDate(input *v1alpha1.DBCluster) (bool, error) {
	panic("implement me")
}
