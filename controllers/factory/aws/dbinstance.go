package aws

import (
	"github.com/agill17/db-operator/api/v1alpha1"
)

func (i InternalAwsClients) CreateDBInstance(input *v1alpha1.DBInstance, password string) error {
	panic("implement me")
}

func (i InternalAwsClients) DeleteDBInstance(input *v1alpha1.DBInstance) error {
	panic("implement me")
}

func (i InternalAwsClients) ModifyDBInstance(input *v1alpha1.DBInstance, password string) error {
	panic("implement me")
}

func (i InternalAwsClients) DBInstanceExists(input *v1alpha1.DBInstance) (bool, error) {
	panic("implement me")
}

func (i InternalAwsClients) IsDBInstanceUpToDate(input *v1alpha1.DBInstance) (bool, error) {
	panic("implement me")
}
