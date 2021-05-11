package v1alpha1

import (
	"fmt"
)

func (in *DBCluster) GetDBClusterID() string {
	clusterID := fmt.Sprintf("%s-%s", in.GetNamespace(), in.GetName())
	if in.Spec.DBClusterIdentifierOverride != "" {
		clusterID = in.Spec.DBClusterIdentifierOverride
	}
	return clusterID
}
