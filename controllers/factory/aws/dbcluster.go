package aws

import "github.com/agill17/db-operator/api/v1alpha1"

func (R RDSClient) CreateDBCluster(in *v1alpha1.DBCluster) error {
	panic("implement me")
}

func (R RDSClient) DeleteDBCluster(in *v1alpha1.DBCluster) error {
	panic("implement me")
}

func (R RDSClient) ModifyDBCluster(in *v1alpha1.DBCluster) error {
	panic("implement me")
}

func (R RDSClient) DBClusterExists(in *v1alpha1.DBCluster) (bool, error) {
	panic("implement me")
}

func (R RDSClient) IsDBClusterUpToDate(in *v1alpha1.DBCluster) (bool, error) {
	panic("implement me")
}
