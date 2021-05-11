package factory

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	internalAwsImpl "github.com/agill17/db-operator/controllers/factory/aws"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type DBCluster interface {
	CreateDBCluster(Obj client.Object, client client.Client, scheme *runtime.Scheme) error
	ModifyDBCluster(Obj client.Object, client client.Client, scheme *runtime.Scheme) error
	IsDBClusterUpToDate(Obj client.Object, client client.Client, scheme *runtime.Scheme) (bool, error)
	DeleteDBCluster(Obj client.Object) error
	DBClusterExists(Obj client.Object) (bool, error)
}

type DBInstance interface {
	CreateDBInstance(Obj client.Object, client client.Client, scheme *runtime.Scheme) error
	DeleteDBInstance(Obj client.Object) error
	ModifyDBInstance(Obj client.Object, client client.Client, scheme *runtime.Scheme) error
	DBInstanceExists(Obj client.Object) (bool, error)
	IsDBInstanceUpToDate(Obj client.Object, client client.Client, scheme *runtime.Scheme) (bool, error)
}

type CloudDB interface {
	DBCluster
	DBInstance
	GetOrSetMasterPassword(Obj client.Object, client client.Client, scheme *runtime.Scheme) (string, error)
}

func NewCloudDB(pType v1alpha1.ProviderType, providerSecret *v1.Secret, region string) (CloudDB, error) {
	if pType == v1alpha1.AWS {
		return internalAwsImpl.NewInternalAwsClient(region, string(pType), providerSecret.Data)
	}

	return nil, errors.New(fmt.Sprintf("Provider %v is not yet supported..", pType))
}
