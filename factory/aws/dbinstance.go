package aws

import "github.com/agill17/db-operator/api/v1alpha1"

func (R RDSClient) CreateDBInstance(in *v1alpha1.DBInstance) error {
	panic("implement me")
}

func (R RDSClient) DeleteDBInstance(in *v1alpha1.DBInstance) error {
	panic("implement me")
}

func (R RDSClient) ModifyDBInstance(in *v1alpha1.DBInstance) error {
	panic("implement me")
}

func (R RDSClient) DBInstanceExists(in *v1alpha1.DBInstance) (bool, error) {
	panic("implement me")
}

func (R RDSClient) IsDBInstanceUpToDate(in *v1alpha1.DBInstance) (bool, error) {
	panic("implement me")
}
