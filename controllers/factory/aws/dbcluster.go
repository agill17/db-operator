package aws

import (
	"errors"
	"fmt"
	"github.com/agill17/db-operator/api/v1alpha1"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/rds"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func clientObjToDBCluster(obj client.Object) (*v1alpha1.DBCluster, error) {
	dbCluster, ok := obj.(*v1alpha1.DBCluster)
	if !ok {
		return nil, errors.New(fmt.Sprintf("ErrCasting%TtoDBCluster", obj))
	}
	return dbCluster, nil
}

func (i InternalAwsClients) CreateDBCluster(input *v1alpha1.DBCluster, password string) error {
	_, errCreating := i.rdsClient.CreateDBCluster(createDBClusterInput(input, password))
	return errCreating
}

func (i InternalAwsClients) DeleteDBCluster(input *v1alpha1.DBCluster) error {
	dbCluster, errCasting := clientObjToDBCluster(input)
	if errCasting != nil {
		return errCasting
	}

	// dont even make a delete attempt if deletionProtection is enabled in CR spec.
	if dbCluster.Spec.DeletionProtection {
		return ErrDBClusterDeletionProtectionEnabled{Message: fmt.Sprintf(
			"%v/%v: Cannot delete, deletion protection is enabled", dbCluster.GetNamespace(),
			dbCluster.GetName())}
	}
	if _, errDeleting := i.rdsClient.DeleteDBCluster(deleteDBClusterInput(dbCluster)); errDeleting != nil {
		if awsErr, isAwsErr := errDeleting.(awserr.Error); isAwsErr {
			if awsErr.Error() == rds.ErrCodeDBClusterNotFoundFault { // if for some reason the dbCluster is not found, ignore and move on
				return nil
			}
		}
		return errDeleting
	}
	return nil
}

func (i InternalAwsClients) ModifyDBCluster(modifyIn interface{}, password string) error {
	rdsModifyIn, ok := modifyIn.(*rds.ModifyDBClusterInput)
	if !ok {
		return errors.New("ErrModifyDBClusterRecived")
	}
	rdsModifyIn.ApplyImmediately = aws.Bool(true)
	if _, errUpdating := i.rdsClient.ModifyDBCluster(rdsModifyIn); errUpdating != nil {
		return errUpdating
	}
	return nil
}

func (i InternalAwsClients) DBClusterExists(dbClusterID string) (bool, string, error) {
	out, err := i.rdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(dbClusterID),
	})
	if err != nil {
		if awsErr, isAwsErr := err.(awserr.Error); isAwsErr {
			if awsErr.Code() == rds.ErrCodeDBClusterNotFoundFault {
				return false, "", nil
			}
		}
		return false, "", err
	}
	return true, *out.DBClusters[0].Status, nil
}

// TODO: refactor, I am not proud of this.. 
func (i InternalAwsClients) IsDBClusterUpToDate(input *v1alpha1.DBCluster) (bool, interface{}, error) {
	clusterState, err := i.rdsClient.DescribeDBClusters(&rds.DescribeDBClustersInput{
		DBClusterIdentifier:aws.String(input.GetDBClusterID())})
	if err != nil {
		return false, nil, err
	}
	if len(clusterState.DBClusters) != 1 {
		return false, nil,errors.New("ErrMultipleDBClustersExistsWithTheSameID");
	}
	currentState := clusterState.DBClusters[0]

	modifyDBClusterInput := &rds.ModifyDBClusterInput{}
	if *currentState.EngineVersion != input.Spec.EngineVersion {
		modifyDBClusterInput.EngineVersion = aws.String(input.Spec.EngineVersion)
	}
	if *currentState.HttpEndpointEnabled != input.Spec.EnableHttpEndpoint {
		modifyDBClusterInput.EnableHttpEndpoint = aws.Bool(input.Spec.EnableHttpEndpoint)
	}
	if *currentState.GlobalWriteForwardingRequested != input.Spec.EnableGlobalWriteForwarding {
		modifyDBClusterInput.EnableGlobalWriteForwarding = aws.Bool(input.Spec.EnableGlobalWriteForwarding)
	}
	if *currentState.DeletionProtection != input.Spec.DeletionProtection {
		modifyDBClusterInput.DeletionProtection = aws.Bool(input.Spec.DeletionProtection)
	}
	if *currentState.CopyTagsToSnapshot != input.Spec.CopyTagsToSnapshot {
		modifyDBClusterInput.CopyTagsToSnapshot = aws.Bool(input.Spec.CopyTagsToSnapshot)
	}
	if *currentState.BackupRetentionPeriod != input.Spec.BackupRetentionPeriod {
		modifyDBClusterInput.BackupRetentionPeriod = aws.Int64(input.Spec.BackupRetentionPeriod)
	}
	if *currentState.DBClusterParameterGroup != input.Spec.DBClusterParameterGroupName {
		modifyDBClusterInput.DBClusterParameterGroupName = aws.String(input.Spec.DBClusterParameterGroupName)
	}
	if *currentState.DBClusterIdentifier != input.GetDBClusterID() {
		modifyDBClusterInput.NewDBClusterIdentifier = aws.String(input.GetDBClusterID())
	}
	if *currentState.Port != input.Spec.Port {
		modifyDBClusterInput.Port = aws.Int64(input.Spec.Port)
	}
	if *currentState.PreferredBackupWindow != input.Spec.PreferredBackupWindow {
		modifyDBClusterInput.PreferredBackupWindow = aws.String(input.Spec.PreferredBackupWindow)
	}
	if *currentState.PreferredMaintenanceWindow != input.Spec.PreferredMaintenanceWindow {
		modifyDBClusterInput.PreferredMaintenanceWindow = aws.String(input.Spec.PreferredMaintenanceWindow)
	}
	if len(currentState.VpcSecurityGroups) != len(input.Spec.VpcSecurityGroupIds) {
		modifyDBClusterInput.VpcSecurityGroupIds = aws.StringSlice(input.Spec.VpcSecurityGroupIds)
	}

	return modifyDBClusterInput == &rds.ModifyDBClusterInput{}, modifyDBClusterInput, nil
}