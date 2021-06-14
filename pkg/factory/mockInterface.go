package factory

import (
	"github.com/agill17/db-operator/api/v1alpha1"
)

type MockCloudDB struct {
	CloudDB
	CreateDBClusterErr          error
	DeleteDBClusterErr          error
	IsDBClusterUpToDateResp     bool
	IsDBClusterUpToDateModifyIn interface{}
	IsDBClusterUpToDateErr      error
	DBStatusResp                *v1alpha1.DBStatus
	DBClusterExistsErr          error
	ModifyDBClusterErr          error
}

func (m *MockCloudDB) CreateDBCluster(input *v1alpha1.DBCluster, password string) error {
	return m.CreateDBClusterErr
}
func (m *MockCloudDB) ModifyDBCluster(modifyIn interface{}) error {
	return m.ModifyDBClusterErr
}
func (m *MockCloudDB) IsDBClusterUpToDate(input *v1alpha1.DBCluster) (bool, interface{}, error) {
	return m.IsDBClusterUpToDateResp, m.IsDBClusterUpToDateModifyIn, m.IsDBClusterUpToDateErr
}
func (m *MockCloudDB) DeleteDBCluster(input *v1alpha1.DBCluster) error {
	return m.DeleteDBClusterErr
}
func (m *MockCloudDB) DBClusterExists(dbClusterID string) (*v1alpha1.DBStatus, error) {
	return m.DBStatusResp, m.DBClusterExistsErr
}

//type MockRDS struct {
//	rdsiface.RDSAPI
//	RdsCreateDBClusterResp *rds.CreateDBClusterOutput
//	RdsCreateDBClusterErr error
//	RdsDeleteDBClusterResp *rds.DeleteDBClusterOutput
//	RdsDeleteDBClusterErr error
//	RdsDescribeDBClusterResp *rds.DescribeDBClustersOutput
//	RdsDescribeDBClusterErr error
//	RdsModifyDBClusterResp *rds.ModifyDBClusterOutput
//	RdsModifyDBClusterErr error
//}
//
//func (m *MockRDS) DescribeDBClusters(in *rds.DescribeDBClustersInput) (*rds.DescribeDBClustersOutput, error) {
//	return m.RdsDescribeDBClusterResp, m.RdsDescribeDBClusterErr
//}
//func (m *MockRDS) DeleteDBCluster(in *rds.DeleteDBClusterInput) (*rds.DeleteDBClusterOutput, error) {
//	return m.RdsDeleteDBClusterResp, m.RdsDeleteDBClusterErr
//}
//func (m *MockRDS) CreateDBCluster(in *rds.CreateDBClusterInput) (*rds.CreateDBClusterOutput, error) {
//	return m.RdsCreateDBClusterResp, m.RdsCreateDBClusterErr
//}
//func (m *MockRDS) ModifyDBCluster(in *rds.ModifyDBClusterInput) (*rds.ModifyDBClusterOutput, error) {
//	return m.RdsModifyDBClusterResp, m.RdsModifyDBClusterErr
//}
