package factory

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	internalAwsImpl "github.com/agill17/db-operator/pkg/factory/aws"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
)

type DBCluster interface {
	CreateDBCluster(input *v1alpha1.DBCluster, password string) error
	ModifyDBCluster(modifyIn interface{}) error
	IsDBClusterUpToDate(input *v1alpha1.DBCluster) (bool, interface{}, error)
	DeleteDBCluster(input *v1alpha1.DBCluster) error
	DBClusterExists(dbClusterID string) (*v1alpha1.DBStatus, error)
}

type DBInstance interface {
	CreateDBInstance(input *v1alpha1.DBInstance, password string) error
	DeleteDBInstance(input *v1alpha1.DBInstance) error
	ModifyDBInstance(modifyIn interface{}) error
	DBInstanceExists(input *v1alpha1.DBInstance) (*v1alpha1.DBStatus, error)
	IsDBInstanceUpToDate(input *v1alpha1.DBInstance) (bool, interface{}, error)
}

type CloudDB interface {
	DBCluster
	DBInstance
}

func NewCloudDB(logger logr.Logger, pType v1alpha1.ProviderType, providerSecret *v1.Secret, region string) (CloudDB, error) {
	if pType == v1alpha1.AWS {
		return internalAwsImpl.NewInternalAwsClient(logger, region, string(pType), providerSecret.Data)
	}

	return nil, errors.New(fmt.Sprintf("Provider %v is not yet supported..", pType))
}
