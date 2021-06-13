/*
Copyright 2021 agill17.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MasterUserPasswordSecretRef struct {
	// The key in secret.data that contains the master password
	PasswordKey string             `json:"passwordKey"`
	SecretRef   v1.SecretReference `json:"secretRef"`
}

// DBClusterSpec defines the desired state of DBCluster
type DBClusterSpec struct {
	Provider Provider `json:"provider,required"`
	Region   string   `json:"region,required"`
	// A list of Availability Zones (AZs) where instances in the DB cluster can
	// be created. For information on AWS Regions and Availability Zones, see Choosing
	// the Regions and Availability Zones (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Concepts.RegionsAndAvailabilityZones.html)
	// in the Amazon Aurora User Guide.
	AvailabilityZones []string `json:"availabilityZones,required"`

	// The number of days for which automated backups are retained.
	//
	// For AWS, Default: 1
	//
	// Constraints:
	//    * Must be a value from 1 to 35
	// +kubebuilder:validation:Minimum:=1
	// +kubebuilder:validation:Maximum:=35
	// +kubebuilder:default:=1
	// +optional
	BackupRetentionPeriod int64 `json:"backupRetentionPeriod,optional"`

	// A value that indicates whether to copy all tags from the DB cluster to snapshots
	// of the DB cluster. The default is not to copy them.
	// +optional
	CopyTagsToSnapshot bool `json:"copyTagsToSnapshot,optional"`

	// The DB cluster identifier. This parameter is stored as a lowercase string.
	// Constraints:
	//    * Must contain from 1 to 63 letters, numbers, or hyphens.
	//    * First character must be a letter.
	//    * Can't end with a hyphen or contain two consecutive hyphens.
	// Example: my-cluster1
	// DBClusterIdentifierOverride is a optional field, defaults to .metadata.name
	// +optional
	DBClusterIdentifierOverride string `json:"dbClusterIdentifierOverride,optional"`

	// The name of the DB cluster parameter group to associate with this DB cluster.
	// If you do not specify a value, then the default DB cluster parameter group
	// for the specified DB engine and version is used.
	// Constraints:
	//    * If supplied, must match the name of an existing DB cluster parameter
	//    group.
	DBClusterParameterGroupName string `json:"dbClusterParameterGroupName,optional"`

	// A DB subnet group to associate with this DB cluster.
	//
	// Constraints: Must match the name of an existing DBSubnetGroup. Must not be
	// default.
	//
	// Example: mySubnetgroup
	// +optional
	DBSubnetGroupName string `json:"dbSubnetGroupName,optional"`

	// The name for your database of up to 64 alphanumeric characters. If you do
	// not provide a name, Amazon RDS doesn't create a database in the DB cluster
	// you are creating.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	DatabaseName string `json:"databaseName,required"`

	// A value that indicates whether the DB cluster has deletion protection enabled.
	// The database can't be deleted when deletion protection is enabled. By default,
	// deletion protection is disabled.
	DeletionProtection bool `json:"deletionProtection,required"`

	// DestinationRegion is used for presigning the request to a given region.
	// +optional
	DestinationRegion string `json:"destinationRegion,optional"`

	// The list of log types that need to be enabled for exporting to CloudWatch
	// Logs. The values in the list depend on the DB engine being used. For more
	// information, see Publishing Database Logs to Amazon CloudWatch Logs (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/USER_LogAccess.html#USER_LogAccess.Procedural.UploadtoCloudWatch)
	// in the Amazon Aurora User Guide.
	//
	// Aurora MySQL
	//
	// Possible values are audit, error, general, and slowquery.
	//
	// Aurora PostgreSQL
	//
	// Possible value is postgresql.
	//TODO: add AnyOf by hand since controller-gen does not have this baked in: https://github.com/kubernetes-sigs/controller-tools/issues/461
	// +optional
	EnableCloudwatchLogsExports []string `json:"enableCouldWatchLogExport,optional"`

	// A value that indicates whether to enable this DB cluster to forward write
	// operations to the primary cluster of an Aurora global database (GlobalCluster).
	// By default, write operations are not allowed on Aurora DB clusters that are
	// secondary clusters in an Aurora global database.
	//
	// You can set this value only on Aurora DB clusters that are members of an
	// Aurora global database. With this parameter enabled, a secondary cluster
	// can forward writes to the current primary cluster and the resulting changes
	// are replicated back to this cluster. For the primary DB cluster of an Aurora
	// global database, this value is used immediately if the primary is demoted
	// by the FailoverGlobalCluster API operation, but it does nothing until then.
	// +optional
	EnableGlobalWriteForwarding bool `json:"enableGlobalWriteForwarding,optional"`

	// A value that indicates whether to enable the HTTP endpoint for an Aurora
	// Serverless DB cluster. By default, the HTTP endpoint is disabled.
	//
	// When enabled, the HTTP endpoint provides a connectionless web service API
	// for running SQL queries on the Aurora Serverless DB cluster. You can also
	// query your database from inside the RDS console with the query editor.
	//
	// For more information, see Using the Data API for Aurora Serverless (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/data-api.html)
	// in the Amazon Aurora User Guide.
	// +optional
	EnableHttpEndpoint bool `json:"enableHttpEndpoint,optional"`

	// The name of the database engine to be used for this DB cluster.
	//
	// Valid Values: aurora (for MySQL 5.6-compatible Aurora), aurora-mysql (for
	// MySQL 5.7-compatible Aurora), and aurora-postgresql
	//
	// Engine is a required field
	// +kubebuilder:validation:Enum:="aurora";"aurora-mysql";"aurora-postgresql"
	Engine string `json:"engine,required"`

	// The DB engine mode of the DB cluster, either provisioned, serverless, parallelquery,
	// global, or multimaster.
	//
	// The parallelquery engine mode isn't required for Aurora MySQL version 1.23
	// and higher 1.x versions, and version 2.09 and higher 2.x versions.
	//
	// The global engine mode isn't required for Aurora MySQL version 1.22 and higher
	// 1.x versions, and global engine mode isn't required for any 2.x versions.
	//
	// The multimaster engine mode only applies for DB clusters created with Aurora
	// MySQL version 5.6.10a.
	//
	// For Aurora PostgreSQL, the global engine mode isn't required, and both the
	// parallelquery and the multimaster engine modes currently aren't supported.
	//
	// Limitations and requirements apply to some DB engine modes. For more information,
	// see the following sections in the Amazon Aurora User Guide:
	//
	//    * Limitations of Aurora Serverless (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-serverless.html#aurora-serverless.limitations)
	//
	//    * Limitations of Parallel Query (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-mysql-parallel-query.html#aurora-mysql-parallel-query-limitations)
	//
	//    * Limitations of Aurora Global Databases (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-global-database.html#aurora-global-database.limitations)
	//
	//    * Limitations of Multi-Master Clusters (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/aurora-multi-master.html#aurora-multi-master-limitations)
	// +kubebuilder:validation:Enum:="provisioned";"serverless";"parallelquery";"global";"multimaster"
	EngineMode string `json:"engineMode,required"`

	// The version number of the database engine to use.
	//
	// To list all of the available engine versions for aurora (for MySQL 5.6-compatible
	// Aurora), use the following command:
	//
	// aws rds describe-db-engine-versions --engine aurora --query "DBEngineVersions[].EngineVersion"
	//
	// To list all of the available engine versions for aurora-mysql (for MySQL
	// 5.7-compatible Aurora), use the following command:
	//
	// aws rds describe-db-engine-versions --engine aurora-mysql --query "DBEngineVersions[].EngineVersion"
	//
	// To list all of the available engine versions for aurora-postgresql, use the
	// following command:
	//
	// aws rds describe-db-engine-versions --engine aurora-postgresql --query "DBEngineVersions[].EngineVersion"
	//
	// Aurora MySQL
	//
	// Example: 5.6.10a, 5.6.mysql_aurora.1.19.2, 5.7.12, 5.7.mysql_aurora.2.04.5
	//
	// Aurora PostgreSQL
	//
	// Example: 9.6.3, 10.7
	EngineVersion string `json:"engineVersion,required"`

	// The AWS KMS key identifier for an encrypted DB cluster.
	//
	// The AWS KMS key identifier is the key ARN, key ID, alias ARN, or alias name
	// for the AWS KMS customer master key (CMK). To use a CMK in a different AWS
	// account, specify the key ARN or alias ARN.
	//
	// When a CMK isn't specified in KmsKeyId:
	//
	//    * If ReplicationSourceIdentifier identifies an encrypted source, then
	//    Amazon RDS will use the CMK used to encrypt the source. Otherwise, Amazon
	//    RDS will use your default CMK.
	//
	//    * If the StorageEncrypted parameter is enabled and ReplicationSourceIdentifier
	//    isn't specified, then Amazon RDS will use your default CMK.
	//
	// There is a default CMK for your AWS account. Your AWS account has a different
	// default CMK for each AWS Region.
	//
	// If you create a read replica of an encrypted DB cluster in another AWS Region,
	// you must set KmsKeyId to a AWS KMS key identifier that is valid in the destination
	// AWS Region. This CMK is used to encrypt the read replica in that AWS Region.
	// +optional
	KmsKeyId string `json:"kmsKeyID,optional"`

	// Specifies the secret to use
	PasswordRef PasswordRef `json:"passwordRef,required"`

	// The name of the master user for the DB cluster.
	// Constraints:
	//    * Must be 1 to 16 letters or numbers.
	//    * First character must be a letter.
	//    * Can't be a reserved word for the chosen database engine.
	MasterUsername string `json:"masterUsername,required"`

	// A value that indicates that the DB cluster should be associated with the
	// specified option group.
	// Permanent options can't be removed from an option group. The option group
	// can't be removed from a DB cluster once it is associated with a DB cluster.
	// +optional
	OptionGroupName string `json:"optionGroupName,optional"`

	// The port number on which the instances in the DB cluster accept connections.
	// Default: 3306 if engine is set as aurora or 5432 if set to aurora-postgresql.
	// +optional
	Port int64 `json:"port,optional"`

	// The daily time range during which automated backups are created if automated
	// backups are enabled using the BackupRetentionPeriod parameter.
	//
	// The default is a 30-minute window selected at random from an 8-hour block
	// of time for each AWS Region. To view the time blocks available, see Backup
	// window (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Managing.Backups.html#Aurora.Managing.Backups.BackupWindow)
	// in the Amazon Aurora User Guide.
	//
	// Constraints:
	//
	//    * Must be in the format hh24:mi-hh24:mi.
	//
	//    * Must be in Universal Coordinated Time (UTC).
	//
	//    * Must not conflict with the preferred maintenance window.
	//
	//    * Must be at least 30 minutes.
	// +optional
	// +kubebuilder:default="23:00-23:30"
	PreferredBackupWindow string `json:"preferredBackupWindow,optional"`

	// The weekly time range during which system maintenance can occur, in Universal
	// Coordinated Time (UTC).
	//
	// Format: ddd:hh24:mi-ddd:hh24:mi
	//
	// The default is a 30-minute window selected at random from an 8-hour block
	// of time for each AWS Region, occurring on a random day of the week. To see
	// the time blocks available, see Adjusting the Preferred DB Cluster Maintenance
	// Window (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/USER_UpgradeDBInstance.Maintenance.html#AdjustingTheMaintenanceWindow.Aurora)
	// in the Amazon Aurora User Guide.
	//
	// Valid Days: Mon, Tue, Wed, Thu, Fri, Sat, Sun.
	//
	// Constraints: Minimum 30-minute window.
	// +optional
	// +kubebuilder:default="sun:06:00-sun:06:30"
	PreferredMaintenanceWindow string `json:"preferredMaintenanceWindow,optional"`

	// The Amazon Resource Name (ARN) of the source DB instance or DB cluster if
	// this DB cluster is created as a read replica.
	// +optional
	ReplicationSourceIdentifier string `json:"replicationSourceIdentifier,optional"`

	// A value that indicates whether the DB cluster is encrypted.
	// +optional
	StorageEncrypted bool `json:"storageEncrypted,optional"`

	// Tags to assign to the DB cluster.
	// +optional
	Tags map[string]string `json:"tags,optional"`

	// A list of EC2 VPC security groups to associate with this DB cluster.
	// +optional
	VpcSecurityGroupIds []string `json:"vpcSecurityGroupIds,optional"`

	// +optional
	// +kubebuilder:default=true
	SkipFinalSnapshot bool `json:"skipFinalSnapshot,optional"`
}

type DBClusterStatus struct {
	Phase                           Phase  `json:"phase"`
	ProviderSecretResourceVersion   string `json:"providerSecretResourceVersion"`
	DBPasswordSecretResourceVersion string `json:"dbPasswordSecretResourceVersion"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
// DBCluster is the Schema for the dbclusters API
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
type DBCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBClusterSpec   `json:"spec,omitempty"`
	Status DBClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBClusterList contains a list of DBCluster
type DBClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBCluster{}, &DBClusterList{})
}
