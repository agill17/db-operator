package factory

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	aws2 "github.com/agill17/db-operator/controllers/factory/aws"
	v1 "k8s.io/api/core/v1"
)

type DBCluster interface {
	CreateDBCluster(in *v1alpha1.DBCluster) error
	DeleteDBCluster(in *v1alpha1.DBCluster) error
	ModifyDBCluster(in *v1alpha1.DBCluster) error
	DBClusterExists(in *v1alpha1.DBCluster) (bool, error)
	IsDBClusterUpToDate(in *v1alpha1.DBCluster) (bool, error)
}

type DBInstance interface {
	CreateDBInstance(in *v1alpha1.DBInstance) error
	DeleteDBInstance(in *v1alpha1.DBInstance) error
	ModifyDBInstance(in *v1alpha1.DBInstance) error
	DBInstanceExists(in *v1alpha1.DBInstance) (bool, error)
	IsDBInstanceUpToDate(in *v1alpha1.DBInstance) (bool, error)
}

type CloudDB interface {
	DBCluster
	DBInstance
}

func NewCloudDB(pType v1alpha1.ProviderType, providerSecret *v1.Secret, region string) (CloudDB, error) {
	if pType == v1alpha1.AWS {
		return aws2.NewRDSClient(region, string(pType), providerSecret.Data)
	}

	return nil, errors.New(fmt.Sprintf("Provider %v is not yet supported..", pType))
}
