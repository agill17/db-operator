package v1alpha1

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
)

// TODO: These functions need to go away from api dir and need to become mapping function per cloud provider under their own factory implementation

func (in *DBCluster) CreateDBClusterInput(password string) *rds.CreateDBClusterInput {
	out := &rds.CreateDBClusterInput{
		DBClusterIdentifier:         aws.String(in.GetDBClusterID()),
		CopyTagsToSnapshot:          aws.Bool(true),
		DatabaseName:                aws.String(in.Spec.DatabaseName),
		DeletionProtection:          aws.Bool(in.Spec.DeletionProtection),
		EnableCloudwatchLogsExports: aws.StringSlice(in.Spec.EnableCloudwatchLogsExports),
		Engine:                      aws.String(in.Spec.Engine),
		EngineMode:                  aws.String(in.Spec.EngineMode),
		EngineVersion:               aws.String(in.Spec.EngineVersion),
		KmsKeyId:                    aws.String(in.Spec.KmsKeyId),
		MasterUserPassword:          aws.String(password),
		MasterUsername:              aws.String(in.Spec.MasterUsername),
		Port:                        aws.Int64(in.Spec.Port),
		PreferredBackupWindow:       aws.String(in.Spec.PreferredBackupWindow),
		PreferredMaintenanceWindow:  aws.String(in.Spec.PreferredMaintenanceWindow),
		ReplicationSourceIdentifier: aws.String(in.Spec.ReplicationSourceIdentifier),
		StorageEncrypted:            aws.Bool(in.Spec.StorageEncrypted),
		Tags:                        []*rds.Tag{{Key: aws.String("foo"), Value: aws.String("bar")}},
		VpcSecurityGroupIds:         aws.StringSlice(in.Spec.VpcSecurityGroupIds),
	}

	if in.Spec.DBClusterIdentifierOverride != "" {
		out.DBClusterIdentifier = aws.String(in.Spec.DBClusterIdentifierOverride)
	}

	return out
}

func (in *DBCluster) DeleteDBClusterInput() *rds.DeleteDBClusterInput {
	return nil
}
func (in *DBCluster) ModifyDBClusterInput(password string) *rds.ModifyDBClusterInput {
	return nil
}

func (in *DBCluster) GetDBClusterID() string {
	clusterID := fmt.Sprintf("%s-%s", in.GetNamespace(), in.GetName())
	if in.Spec.DBClusterIdentifierOverride != "" {
		clusterID = in.Spec.DBClusterIdentifierOverride
	}
	return clusterID
}
