package factory

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/agill17/db-operator/factory/aws"
)

type DBCluster interface {
	CreateDBCluster() error
	DeleteDBCluster() error
	ModifyDBCluster() error
	DBClusterExists() (bool, error)
	IsDBClusterUpToDate() (bool, error)
}

type DBInstance interface {
	CreateDBInstance() error
	DeleteDBInstance() error
	ModifyDBInstance() error
	DBInstanceExists() (bool, error)
	IsDBInstanceUpToDate() (bool, error)
}

type CloudDB interface {
	DBCluster
	DBInstance
}


func NewCloudDB(provider *v1alpha1.Provider, region string) (CloudDB, error){
	pType := provider.Spec.Type
	if pType == v1alpha1.AWS {
		return aws.NewRDSClient(region, provider.GetName(), provider.Spec.Credentials)
	}

	return nil, errors.New(fmt.Sprintf("Provider %v is not yet supported..", pType))
}