package aws

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func clientObjToDBCluster(obj client.Object) (*v1alpha1.DBCluster, error) {
	dbCluster, ok := obj.(*v1alpha1.DBCluster)
	if !ok {
		return nil, errors.New(fmt.Sprintf("ErrCasting%TtoDBCluster", obj))
	}
	return dbCluster, nil
}

func (i InternalAwsClients) CreateDBCluster(Obj client.Object, client client.Client, scheme *runtime.Scheme) error {
	dbCluster, errCasting := clientObjToDBCluster(Obj)
	if errCasting != nil {
		return errCasting
	}
	pass, err := i.GetOrSetMasterPassword(Obj, client, scheme)
	if err != nil {
		return err
	}
	_, errCreating := i.rdsClient.CreateDBCluster(createDBClusterInput(dbCluster, pass))
	return errCreating
}

func (i InternalAwsClients) DeleteDBCluster(Obj client.Object) error {
	dbCluster, errCasting := clientObjToDBCluster(Obj)
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

func (i InternalAwsClients) ModifyDBCluster(Obj client.Object, client client.Client, scheme *runtime.Scheme) error {
	dbCluster, errCasting := clientObjToDBCluster(Obj)
	if errCasting != nil {
		return errCasting
	}
	pass, err := i.GetOrSetMasterPassword(Obj, client, scheme)
	if err != nil {
		return err
	}
	if _, errUpdating := i.rdsClient.ModifyDBCluster(modifyDBClusterInput(dbCluster, pass)); errUpdating != nil {
		return errUpdating
	}
	return nil
}

func (i InternalAwsClients) DBClusterExists(Obj client.Object) (bool, error) {
	dbCluster, errCasting := clientObjToDBCluster(Obj)
	if errCasting != nil {
		return false, errCasting
	}
	_, err := i.rdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(dbCluster.GetDBClusterID()),
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

func (i InternalAwsClients) IsDBClusterUpToDate(Obj client.Object, client client.Client, scheme *runtime.Scheme) (bool, error) {
	panic("implement me")
}
