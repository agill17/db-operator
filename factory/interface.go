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


func NewDBClusterInterface(providerType v1alpha1.ProviderType, region string, providerCredentialsMap map[string][]byte) (CloudDB, error){
	switch providerType {
	case v1alpha1.AWS:
		return aws.NewRDSClient(region, providerCredentialsMap)
	}
	return nil, errors.New(fmt.Sprintf("Provider %v is not yet supported..", providerType))
}