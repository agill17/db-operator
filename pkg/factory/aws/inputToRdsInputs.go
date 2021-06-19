package aws

import (
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"strconv"
	"strings"
	"time"
)

func mapToRdsTags(t map[string]string) []*rds.Tag {
	var out []*rds.Tag
	for k, v := range t {
		out = append(out, &rds.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		})
	}
	return out
}

func createDBClusterInput(in *v1alpha1.DBCluster, password string) *rds.CreateDBClusterInput {
	out := &rds.CreateDBClusterInput{
		DBClusterIdentifier:         aws.String(in.GetDBClusterID()),
		CopyTagsToSnapshot:          aws.Bool(in.Spec.CopyTagsToSnapshot),
		DatabaseName:                aws.String(in.Spec.DatabaseName),
		DeletionProtection:          aws.Bool(in.Spec.DeletionProtection),
		EnableCloudwatchLogsExports: aws.StringSlice(in.Spec.EnableCloudwatchLogsExports),
		Engine:                      aws.String(in.Spec.Engine),
		EngineMode:                  aws.String(in.Spec.EngineMode),
		EngineVersion:               aws.String(in.Spec.EngineVersion),
		KmsKeyId:                    aws.String(in.Spec.KmsKeyId),
		MasterUserPassword:          aws.String(password),
		MasterUsername:              aws.String(in.Spec.MasterUsername),
		ReplicationSourceIdentifier: aws.String(in.Spec.ReplicationSourceIdentifier),
		StorageEncrypted:            aws.Bool(in.Spec.StorageEncrypted),
		Tags:                        mapToRdsTags(in.Spec.Tags),
		VpcSecurityGroupIds:         aws.StringSlice(in.Spec.VpcSecurityGroupIds),
	}
	if in.Spec.Port != 0 {
		out.Port = aws.Int64(in.Spec.Port)
	}
	if in.Spec.PreferredMaintenanceWindow != "" {
		out.PreferredMaintenanceWindow = aws.String(in.Spec.PreferredMaintenanceWindow)
	}
	if in.Spec.PreferredBackupWindow != "" {
		out.PreferredBackupWindow = aws.String(in.Spec.PreferredBackupWindow)
	}

	return out
}

func deleteDBClusterInput(in *v1alpha1.DBCluster) *rds.DeleteDBClusterInput {
	timeNow := time.Now().Unix()
	timeNowStr := strconv.FormatInt(timeNow, 10)
	out := &rds.DeleteDBClusterInput{
		DBClusterIdentifier: aws.String(in.GetDBClusterID()),
		SkipFinalSnapshot:   aws.Bool(in.Spec.SkipFinalSnapshot),
	}
	if !in.Spec.SkipFinalSnapshot {
		out.FinalDBSnapshotIdentifier = aws.String(fmt.Sprintf("%s-%s", in.GetDBClusterID(), timeNowStr))
	}
	return out
}

func createInstanceInput(in *v1alpha1.DBInstance, password string) *rds.CreateDBInstanceInput {
	out := &rds.CreateDBInstanceInput{
		AllocatedStorage:            aws.Int64(in.Spec.AllocatedStorage),
		AutoMinorVersionUpgrade:     aws.Bool(in.Spec.AutoMinorVersionUpgrade),
		AvailabilityZone:            aws.String(in.Spec.AvailabilityZone),
		BackupRetentionPeriod:       aws.Int64(in.Spec.BackupRetentionPeriod),
		CopyTagsToSnapshot:          aws.Bool(in.Spec.CopyTagsToSnapshot),
		DBClusterIdentifier:         aws.String(in.Spec.DBClusterID),
		DBInstanceClass:             aws.String(in.Spec.DBInstanceClass),
		DBInstanceIdentifier:        aws.String(in.GetDBInstanceID()),
		DBName:                      aws.String(in.Spec.DBName),
		DBSecurityGroups:            aws.StringSlice(in.Spec.DBSecurityGroups),
		EnableCloudwatchLogsExports: aws.StringSlice(in.Spec.CloudwatchLogsExports),
		EnablePerformanceInsights:   aws.Bool(in.Spec.EnablePerformanceInsights),
		Engine:                      aws.String(in.Spec.Engine),
		EngineVersion:               aws.String(in.Spec.EngineVersion),
		Iops:                        aws.Int64(in.Spec.Iops),
		MonitoringInterval:          aws.Int64(in.Spec.MonitoringInterval),
		MonitoringRoleArn:           aws.String(in.Spec.MonitoringRoleArn),
		MultiAZ:                     aws.Bool(in.Spec.MultiAZ),
		OptionGroupName:             aws.String(in.Spec.OptionGroupName),
		PubliclyAccessible:          aws.Bool(in.Spec.PubliclyAccessible),
		StorageEncrypted:            aws.Bool(in.Spec.StorageEncrypted),
		StorageType:                 aws.String(in.Spec.StorageType),
		VpcSecurityGroupIds:         aws.StringSlice(in.Spec.VpcSecurityGroupIds),
		Tags:                        mapToRdsTags(in.Spec.Tags),
	}
	if in.Spec.Port != 0 {
		out.Port = aws.Int64(in.Spec.Port)
	}
	if in.Spec.DBSubnetGroupName != "" {
		out.DBSubnetGroupName = aws.String(in.Spec.DBSubnetGroupName)
	}
	if in.Spec.DBParameterGroupName != "" {
		out.DBParameterGroupName = aws.String(in.Spec.DBParameterGroupName)
	}
	if in.Spec.StorageEncrypted {
		out.KmsKeyId = aws.String(in.Spec.KmsKeyId)
	}
	if in.Spec.LicenseModel != "" {
		out.LicenseModel = aws.String(in.Spec.LicenseModel)
	}
	if in.Spec.PerformanceInsightsKMSKeyId != "" {
		out.PerformanceInsightsKMSKeyId = aws.String(in.Spec.PerformanceInsightsKMSKeyId)
	}

	// instances that are not part of dnlcuster ( non-aurora for aws )
	if in.Spec.DBClusterID == "" {
		out.MasterUsername = aws.String(in.Spec.MasterUsername)
		out.MasterUserPassword = aws.String(password)
		out.DeletionProtection = aws.Bool(in.Spec.DeletionProtection)
	}

	if strings.HasPrefix(in.Spec.Engine, "sqlserver") && in.Spec.Timezone != "" {
		out.Timezone = aws.String(in.Spec.Timezone)
	}
	return out
}

func deleteDbInstanceInput(instance *v1alpha1.DBInstance) *rds.DeleteDBInstanceInput {
	timeNow := time.Now().Unix()
	timeNowStr := strconv.FormatInt(timeNow, 10)
	out := &rds.DeleteDBInstanceInput{
		DBInstanceIdentifier:   aws.String(instance.GetDBInstanceID()),
		DeleteAutomatedBackups: aws.Bool(false),
		SkipFinalSnapshot:      aws.Bool(instance.Spec.SkipFinalSnapshot),
	}
	if !*out.SkipFinalSnapshot {
		out.FinalDBSnapshotIdentifier = aws.String(fmt.Sprintf("%s-%s", instance.GetDBInstanceID(), timeNowStr))
	}
	return out
}
