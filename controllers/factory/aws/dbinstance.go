package aws

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (i InternalAwsClients) CreateDBInstance(Obj client.Object, client client.Client, scheme *runtime.Scheme) error {
	panic("implement me")
}

func (i InternalAwsClients) DeleteDBInstance(Obj client.Object) error {
	panic("implement me")
}

func (i InternalAwsClients) ModifyDBInstance(Obj client.Object, client client.Client, scheme *runtime.Scheme) error {
	panic("implement me")
}

func (i InternalAwsClients) DBInstanceExists(Obj client.Object) (bool, error) {
	panic("implement me")
}

func (i InternalAwsClients) IsDBInstanceUpToDate(Obj client.Object, client client.Client, scheme *runtime.Scheme) (bool, error) {
	panic("implement me")
}
