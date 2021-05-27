package aws

import (
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"strconv"
	"time"
)

func createDBClusterInput(in *v1alpha1.DBCluster, password string) *rds.CreateDBClusterInput {
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

func deleteDBClusterInput(in *v1alpha1.DBCluster) *rds.DeleteDBClusterInput {
	timeNow := time.Now().Unix()
	timeNowStr := strconv.FormatInt(timeNow, 10)
	return &rds.DeleteDBClusterInput{
		DBClusterIdentifier:       aws.String(in.GetDBClusterID()),
		FinalDBSnapshotIdentifier: aws.String(fmt.Sprintf("%s-%s", in.GetDBClusterID(), timeNowStr)),
		SkipFinalSnapshot:         aws.Bool(!in.Spec.SkipFinalSnapshot),
	}
}
func modifyDBClusterInput(in *v1alpha1.DBCluster, password string) *rds.ModifyDBClusterInput {
	return nil
}
