package factory

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	internalAwsImpl "github.com/agill17/db-operator/controllers/factory/aws"
	v1 "k8s.io/api/core/v1"
)

type DBCluster interface {
	CreateDBCluster(input *v1alpha1.DBCluster, password string) error
	ModifyDBCluster(input *v1alpha1.DBCluster, password string) error
	IsDBClusterUpToDate(input *v1alpha1.DBCluster) (bool, error)
	DeleteDBCluster(input *v1alpha1.DBCluster) error
	DBClusterExists(input *v1alpha1.DBCluster) (bool, error)
}

type DBInstance interface {
	CreateDBInstance(input *v1alpha1.DBInstance, password string) error
	DeleteDBInstance(input *v1alpha1.DBInstance) error
	ModifyDBInstance(input *v1alpha1.DBInstance, password string) error
	DBInstanceExists(input *v1alpha1.DBInstance) (bool, error)
	IsDBInstanceUpToDate(input *v1alpha1.DBInstance) (bool, error)
}

type CloudDB interface {
	DBCluster
	DBInstance
}

func NewCloudDB(pType v1alpha1.ProviderType, providerSecret *v1.Secret, region string) (CloudDB, error) {
	if pType == v1alpha1.AWS {
		return internalAwsImpl.NewInternalAwsClient(region, string(pType), providerSecret.Data)
	}

	return nil, errors.New(fmt.Sprintf("Provider %v is not yet supported..", pType))
}
