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
	"fmt"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DBInstanceSpec defines the desired state of DBInstance
type DBInstanceSpec struct {
	Region   string   `json:"region,required"`
	Provider Provider `json:"provider,required"`
	// +optional
	// +kubebuilder:default=true
	SkipFinalSnapshot bool `json:"skipFinalSnapshot,omitempty"`
	// Applicable for AWS aurora db clusters
	// +optional
	DBClusterID string `json:"dbClusterID,omitempty" required-for-engines:"aurora,aurora-mysql,aurora-postgresql"`
	// The amount of storage (in gibibytes) to allocate for the DB instance.
	// Type: Integer
	// Amazon Aurora
	// Not applicable. Aurora cluster volumes automatically grow as the amount of
	// data in your database increases, though you are only charged for the space
	// that you use in an Aurora cluster volume.
	// MySQL
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Must be an integer from 20 to 65536.
	//    * Provisioned IOPS storage (io1): Must be an integer from 100 to 65536.
	//    * Magnetic storage (standard): Must be an integer from 5 to 3072.
	// MariaDB
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Must be an integer from 20 to 65536.
	//    * Provisioned IOPS storage (io1): Must be an integer from 100 to 65536.
	//    * Magnetic storage (standard): Must be an integer from 5 to 3072.
	// PostgreSQL
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Must be an integer from 20 to 65536.
	//    * Provisioned IOPS storage (io1): Must be an integer from 100 to 65536.
	//    * Magnetic storage (standard): Must be an integer from 5 to 3072.
	// Oracle
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Must be an integer from 20 to 65536.
	//    * Provisioned IOPS storage (io1): Must be an integer from 100 to 65536.
	//    * Magnetic storage (standard): Must be an integer from 10 to 3072.
	// SQL Server
	// Constraints to the amount of storage for each storage type are the following:
	//    * General Purpose (SSD) storage (gp2): Enterprise and Standard editions:
	//    Must be an integer from 200 to 16384. Web and Express editions: Must be
	//    an integer from 20 to 16384.
	//    * Provisioned IOPS storage (io1): Enterprise and Standard editions: Must
	//    be an integer from 200 to 16384. Web and Express editions: Must be an
	//    integer from 100 to 16384.
	//    * Magnetic storage (standard): Enterprise and Standard editions: Must
	//    be an integer from 200 to 1024. Web and Express editions: Must be an integer
	//    from 20 to 1024.
	// required for non-aurora database instances
	AllocatedStorage int64 `json:"allocatedStorage,omitempty" required-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// A value that indicates whether minor engine upgrades are applied automatically
	// to the DB instance during the maintenance window. By default, minor engine
	// upgrades are applied automatically.
	// +optional
	AutoMinorVersionUpgrade bool `json:"autoMinorVersionUpgrade,omitempty"`

	// The Availability Zone (AZ) where the database will be created. For information
	// on AWS Regions and Availability Zones, see Regions and Availability Zones
	// (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.RegionsAndAvailabilityZones.html).
	// Default: A random, system-chosen Availability Zone in the endpoint's AWS
	// Region.
	// Example: us-east-1d
	// Constraint: The AvailabilityZone parameter can't be specified if the DB instance
	// is a Multi-AZ deployment. The specified Availability Zone must be in the
	// same AWS Region as the current endpoint.
	// If you're creating a DB instance in an RDS on VMware environment, specify
	// the identifier of the custom Availability Zone to create the DB instance
	// in.
	// For more information about RDS on VMware, see the RDS on VMware User Guide.
	// (https://docs.aws.amazon.com/AmazonRDS/latest/RDSonVMwareUserGuide/rds-on-vmware.html)
	// +optional
	AvailabilityZone string `json:"availabilityZone,omitempty"`

	// The number of days for which automated backups are retained. Setting this
	// parameter to a positive number enables backups. Setting this parameter to
	// 0 disables automated backups.
	// Amazon Aurora
	// Not applicable. The retention period for automated backups is managed by
	// the DB cluster.
	// Default: 1
	// Constraints:
	//    * Must be a value from 0 to 35
	//    * Can't be set to 0 if the DB instance is a source to read replicas
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=35
	BackupRetentionPeriod int64 `json:"backupRetentionPeriod,omitempty" applicable-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// A value that indicates whether to copy tags from the DB instance to snapshots
	// of the DB instance. By default, tags are not copied.
	// Amazon Aurora
	// Not applicable. Copying tags to snapshots is managed by the DB cluster. Setting
	// this value for an Aurora DB instance has no effect on the DB cluster setting.
	// +optional
	// +kubebuilder:default=true
	CopyTagsToSnapshot bool `json:"copyTagsToSnapshot,omitempty" applicable-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// The compute and memory capacity of the DB instance, for example, db.m4.large.
	// Not all DB instance classes are available in all AWS Regions, or for all
	// database engines. For the full list of DB instance classes, and availability
	// for your engine, see DB Instance Class (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Concepts.DBInstanceClass.html)
	// in the Amazon RDS User Guide.
	//
	// DBInstanceClass is a required field
	DBInstanceClass string `json:"dbInstanceClass,required"`

	// The DB instance identifier. This parameter is stored as a lowercase string.
	// Constraints:
	//    * Must contain from 1 to 63 letters, numbers, or hyphens.
	//    * First character must be a letter.
	//    * Can't end with a hyphen or contain two consecutive hyphens.
	// Example: mydbinstance
	// DBInstanceIdentifier is an optional field, defaults to metadata.namespace-metadata.name
	// +optional
	DBInstanceIdentifierOverride string `json:"dbInstanceIdentifierOverride,omitempty"`

	// The meaning of this parameter differs according to the database engine you
	// use.
	// MySQL
	// The name of the database to create when the DB instance is created. If this
	// parameter isn't specified, no database is created in the DB instance.
	// Constraints:
	//    * Must contain 1 to 64 letters or numbers.
	//    * Must begin with a letter. Subsequent characters can be letters, underscores,
	//    or digits (0-9).
	//    * Can't be a word reserved by the specified database engine
	// MariaDB
	// The name of the database to create when the DB instance is created. If this
	// parameter isn't specified, no database is created in the DB instance.
	// Constraints:
	//    * Must contain 1 to 64 letters or numbers.
	//    * Must begin with a letter. Subsequent characters can be letters, underscores,
	//    or digits (0-9).
	//    * Can't be a word reserved by the specified database engine
	// PostgreSQL
	// The name of the database to create when the DB instance is created. If this
	// parameter isn't specified, a database named postgres is created in the DB
	// instance.
	// Constraints:
	//    * Must contain 1 to 63 letters, numbers, or underscores.
	//    * Must begin with a letter. Subsequent characters can be letters, underscores,
	//    or digits (0-9).
	//    * Can't be a word reserved by the specified database engine
	// Oracle
	// The Oracle System ID (SID) of the created DB instance. If you specify null,
	// the default value ORCL is used. You can't specify the string NULL, or any
	// other reserved word, for DBName.
	// Default: ORCL
	// Constraints:
	//    * Can't be longer than 8 characters
	// SQL Server
	// Not applicable. Must be null.
	// Amazon Aurora MySQL
	// The name of the database to create when the primary DB instance of the Aurora
	// MySQL DB cluster is created. If this parameter isn't specified for an Aurora
	// MySQL DB cluster, no database is created in the DB cluster.
	// Constraints:
	//    * It must contain 1 to 64 alphanumeric characters.
	//    * It can't be a word reserved by the database engine.
	// Amazon Aurora PostgreSQL
	// The name of the database to create when the primary DB instance of the Aurora
	// PostgreSQL DB cluster is created. If this parameter isn't specified for an
	// Aurora PostgreSQL DB cluster, a database named postgres is created in the
	// DB cluster.
	// Constraints:
	//    * It must contain 1 to 63 alphanumeric characters.
	//    * It must begin with a letter or an underscore. Subsequent characters
	//    can be letters, underscores, or digits (0 to 9).
	//    * It can't be a word reserved by the database engine.
	// +optional
	DBName string `json:"dbName,omitempty" applicable-for-engines:"aurora,aurora-mysql,aurora-postgresql,mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres"`

	// The name of the DB parameter group to associate with this DB instance. If
	// you do not specify a value, then the default DB parameter group for the specified
	// DB engine and version is used.
	// Constraints:
	//    * Must be 1 to 255 letters, numbers, or hyphens.
	//    * First character must be a letter
	//    * Can't end with a hyphen or contain two consecutive hyphens
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	DBParameterGroupName string `json:"dbParameterGroupName,omitempty"`

	// A list of DB security groups to associate with this DB instance.
	// Default: The default DB security group for the database engine.
	// +optional
	DBSecurityGroups []string `json:"dbSecurityGroups,omitempty"`

	// A DB subnet group to associate with this DB instance.
	// If there is no DB subnet group, then it is a non-VPC DB instance.
	// +optional
	DBSubnetGroupName string `json:"dbSubnetGroupName,omitempty"`

	// A value that indicates whether the DB instance has deletion protection enabled.
	// The database can't be deleted when deletion protection is enabled. By default,
	// deletion protection is disabled. For more information, see Deleting a DB
	// Instance (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_DeleteInstance.html).
	// Amazon Aurora
	// Not applicable. You can enable or disable deletion protection for the DB
	// cluster. For more information, see CreateDBCluster. DB instances in a DB
	// cluster can be deleted even when deletion protection is enabled for the DB
	// cluster.
	// +optional
	DeletionProtection bool `json:"deletionProtection,omitempty" applicable-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// The list of log types that need to be enabled for exporting to CloudWatch
	// Logs. The values in the list depend on the DB engine being used. For more
	// information, see Publishing Database Logs to Amazon CloudWatch Logs (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_LogAccess.html#USER_LogAccess.Procedural.UploadtoCloudWatch)
	// in the Amazon Relational Database Service User Guide.
	// Amazon Aurora
	// Not applicable. CloudWatch Logs exports are managed by the DB cluster.
	//
	// MariaDB
	// Possible values are audit, error, general, and slowquery.
	//
	// Microsoft SQL Server
	// Possible values are agent and error.
	//
	// MySQL
	// Possible values are audit, error, general, and slowquery.
	//
	// Oracle
	// Possible values are alert, audit, listener, trace, and oemagent.
	//
	// PostgreSQL
	// Possible values are postgresql and upgrade.
	// +optional
	CloudwatchLogsExports []string `json:"cloudwatchLogsExports" applicable-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// A value that indicates whether to enable Performance Insights for the DB
	// instance.
	// For more information, see Using Amazon Performance Insights (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_PerfInsights.html)
	// in the Amazon Relational Database Service User Guide.
	// +optional
	// +kubebuilder:default=false
	EnablePerformanceInsights bool `json:"enablePerformanceInsights,omitempty"`

	// The name of the database engine to be used for this instance.
	// Not every database engine is available for every AWS Region.
	// Valid Values:
	//    * aurora (for MySQL 5.6-compatible Aurora)
	//    * aurora-mysql (for MySQL 5.7-compatible Aurora)
	//    * aurora-postgresql
	//    * mariadb
	//    * mysql
	//    * oracle-ee
	//    * oracle-se2
	//    * oracle-se1
	//    * oracle-se
	//    * postgres
	//    * sqlserver-ee
	//    * sqlserver-se
	//    * sqlserver-ex
	//    * sqlserver-web
	// Engine is a required field
	// +kubebuilder:validation:Enum=aurora;aurora-mysql;aurora-postgresql;mariadb;mysql;oracle-ee;oracle-se2;oracle-se1;postgres;sqlserver-ee;sqlserver-se;sqlserver-ex;sqlserver-web
	Engine string `json:"engine,required"`

	// The version number of the database engine to use.
	// For a list of valid engine versions, use the DescribeDBEngineVersions action.
	// The following are the database engines and links to information about the
	// major and minor versions that are available with Amazon RDS. Not every database
	// engine is available for every AWS Region.
	//
	// Amazon Aurora
	// Not applicable. The version number of the database engine to be used by the
	// DB instance is managed by the DB cluster.
	//
	// MariaDB
	// See MariaDB on Amazon RDS Versions (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_MariaDB.html#MariaDB.Concepts.VersionMgmt)
	// in the Amazon RDS User Guide.
	//
	// Microsoft SQL Server
	// See Microsoft SQL Server Versions on Amazon RDS (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_SQLServer.html#SQLServer.Concepts.General.VersionSupport)
	// in the Amazon RDS User Guide.
	//
	// MySQL
	// See MySQL on Amazon RDS Versions (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_MySQL.html#MySQL.Concepts.VersionMgmt)
	// in the Amazon RDS User Guide.
	//
	// Oracle
	// See Oracle Database Engine Release Notes (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.Oracle.PatchComposition.html)
	// in the Amazon RDS User Guide.
	//
	// PostgreSQL
	// See Amazon RDS for PostgreSQL versions and extensions (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html#PostgreSQL.Concepts)
	// in the Amazon RDS User Guide.
	// required for non-aurora dbs
	// +optional
	EngineVersion string `json:"engineVersion,omitempty" required-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// The amount of Provisioned IOPS (input/output operations per second) to be
	// initially allocated for the DB instance. For information about valid Iops
	// values, see Amazon RDS Provisioned IOPS Storage to Improve Performance (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_Storage.html#USER_PIOPS)
	// in the Amazon RDS User Guide.
	// Constraints: For MariaDB, MySQL, Oracle, and PostgreSQL DB instances, must
	// be a multiple between .5 and 50 of the storage amount for the DB instance.
	// For SQL Server DB instances, must be a multiple between 1 and 50 of the storage
	// amount for the DB instance.
	// +optional
	Iops int64 `json:"iops,omitempty"`

	// The AWS KMS key identifier for an encrypted DB instance.
	// The AWS KMS key identifier is the key ARN, key ID, alias ARN, or alias name
	// for the AWS KMS customer master key (CMK). To use a CMK in a different AWS
	// account, specify the key ARN or alias ARN.
	//
	// Amazon Aurora
	// Not applicable. The AWS KMS key identifier is managed by the DB cluster.
	// For more information, see CreateDBCluster.
	//
	// If StorageEncrypted is enabled, and you do not specify a value for the KmsKeyId
	// parameter, then Amazon RDS uses your default CMK. There is a default CMK
	// for your AWS account. Your AWS account has a different default CMK for each
	// AWS Region.
	// +optional
	KmsKeyId string `json:"kmsKeyID,omitempty" applicable-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// License model information for this DB instance.
	// Valid values: license-included | bring-your-own-license | general-public-license
	// +optional
	LicenseModel string `json:"licenseModel,omitempty"`

	// The password for the master user. The password can include any printable
	// ASCII character except "/", """, or "@".
	//
	// Amazon Aurora
	// Not applicable. The password for the master user is managed by the DB cluster.
	//
	// MariaDB
	// Constraints: Must contain from 8 to 41 characters.
	//
	// Microsoft SQL Server
	// Constraints: Must contain from 8 to 128 characters.
	//
	// MySQL
	// Constraints: Must contain from 8 to 41 characters.
	//
	// Oracle
	// Constraints: Must contain from 8 to 30 characters.
	//
	// PostgreSQL
	// Constraints: Must contain from 8 to 128 characters.
	// required for non-aurora dbs
	// +optional
	PasswordRef PasswordRef `json:"passwordRef,omitempty" required-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// The name for the master user.
	// Amazon Aurora
	// Not applicable. The name for the master user is managed by the DB cluster.
	//
	// MariaDB
	// Constraints:
	//    * Required for MariaDB.
	//    * Must be 1 to 16 letters or numbers.
	//    * Can't be a reserved word for the chosen database engine.
	//
	// Microsoft SQL Server
	// Constraints:
	//    * Required for SQL Server.
	//    * Must be 1 to 128 letters or numbers.
	//    * The first character must be a letter.
	//    * Can't be a reserved word for the chosen database engine.
	//
	// MySQL
	// Constraints:
	//    * Required for MySQL.
	//    * Must be 1 to 16 letters or numbers.
	//    * First character must be a letter.
	//    * Can't be a reserved word for the chosen database engine.
	//
	// Oracle
	// Constraints:
	//    * Required for Oracle.
	//    * Must be 1 to 30 letters or numbers.
	//    * First character must be a letter.
	//    * Can't be a reserved word for the chosen database engine.
	//
	// PostgreSQL
	// Constraints:
	//    * Required for PostgreSQL.
	//    * Must be 1 to 63 letters or numbers.
	//    * First character must be a letter.
	//    * Can't be a reserved word for the chosen database engine.
	// required for non-aurora dbs
	MasterUsername string `json:"masterUsername,omitempty" required-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// The interval, in seconds, between points when Enhanced Monitoring metrics
	// are collected for the DB instance. To disable collecting Enhanced Monitoring
	// metrics, specify 0. The default is 0.
	// If MonitoringRoleArn is specified, then you must also set MonitoringInterval
	// to a value other than 0.
	// Valid Values: 0, 1, 5, 10, 15, 30, 60
	// +optional
	// +kubebuilder:validation:Enum=0;1;5;10;15;30;60
	MonitoringInterval int64 `json:"monitoringInterval,omitempty"`

	// The ARN for the IAM role that permits RDS to send enhanced monitoring metrics
	// to Amazon CloudWatch Logs. For example, arn:aws:iam:123456789012:role/emaccess.
	// For information on creating a monitoring role, go to Setting Up and Enabling
	// Enhanced Monitoring (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Monitoring.OS.html#USER_Monitoring.OS.Enabling)
	// in the Amazon RDS User Guide.
	// If MonitoringInterval is set to a value other than 0, then you must supply
	// a MonitoringRoleArn value.
	// +optional
	MonitoringRoleArn string `json:"monitoringRoleArn"`

	// A value that indicates whether the DB instance is a Multi-AZ deployment.
	// You can't set the AvailabilityZone parameter if the DB instance is a Multi-AZ
	// deployment.
	// +optional
	MultiAZ bool `json:"multiAZ,omitempty"`

	// A value that indicates that the DB instance should be associated with the
	// specified option group.
	// Permanent options, such as the TDE option for Oracle Advanced Security TDE,
	// can't be removed from an option group. Also, that option group can't be removed
	// from a DB instance once it is associated with a DB instance
	// +optional
	OptionGroupName string `json:"optionGroupName,omitempty"`

	// The AWS KMS key identifier for encryption of Performance Insights data.
	// The AWS KMS key identifier is the key ARN, key ID, alias ARN, or alias name
	// for the AWS KMS customer master key (CMK).
	// If you do not specify a value for PerformanceInsightsKMSKeyId, then Amazon
	// RDS uses your default CMK. There is a default CMK for your AWS account. Your
	// AWS account has a different default CMK for each AWS Region.
	// +optional
	PerformanceInsightsKMSKeyId string `json:"performanceInsightsKmsKeyID,omitempty"`

	// The port number on which the database accepts connections.
	// MySQL
	// Default: 3306
	// Valid values: 1150-65535
	// Type: Integer
	//
	// MariaDB
	// Default: 3306
	// Valid values: 1150-65535
	// Type: Integer
	//
	// PostgreSQL
	// Default: 5432
	// Valid values: 1150-65535
	// Type: Integer
	//
	// Oracle
	// Default: 1521
	// Valid values: 1150-65535
	//
	// SQL Server
	// Default: 1433
	// Valid values: 1150-65535 except 1234, 1434, 3260, 3343, 3389, 47001, and
	// 49152-49156.
	//
	// Amazon Aurora
	// Default: 3306
	// Valid values: 1150-65535
	// +optional
	Port int64 `json:"port,omitempty"`

	// +optional
	// +kubebuilder:default=false
	PubliclyAccessible bool `json:"publiclyAccessible,omitempty"`

	// A value that indicates whether the DB instance is encrypted. By default,
	// it isn't encrypted.
	// Amazon Aurora
	// Not applicable. The encryption for DB instances is managed by the DB cluster.
	// +optional
	StorageEncrypted bool `json:"storageEncrypted,omitempty" applicable-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// Specifies the storage type to be associated with the DB instance.
	// Valid values: standard | gp2 | io1
	// If you specify io1, you must also include a value for the Iops parameter.
	// Default: io1 if the Iops parameter is specified, otherwise gp2
	// +kubebuilder:validation:Enum=standard;gp2;io1
	// +optional
	StorageType string `json:"storageType,omitempty"`

	// Tags to assign to the DB instance.
	// +optional
	Tags map[string]string `json:"tags,omitempty"`

	// The time zone of the DB instance. The time zone parameter is currently supported
	// only by Microsoft SQL Server (https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_SQLServer.html#SQLServer.Concepts.General.TimeZone).
	// +optional
	Timezone string `json:"timezone,omitempty" applicable-for-engines:"sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`

	// A list of Amazon EC2 VPC security groups to associate with this DB instance.
	// Amazon Aurora
	// Not applicable. The associated list of EC2 VPC security groups is managed
	// by the DB cluster.
	// Default: The default EC2 VPC security group for the DB subnet group's VPC.
	// +optional
	VpcSecurityGroupIds []string `json:"vpcSecurityGroupIDs,omitempty" applicable-for-engines:"mariadb,mysql,oracle-ee,oracle-se2,oracle-se1,postgres,sqlserver-ee,sqlserver-se,sqlserver-ex,sqlserver-web"`
}

// DBInstanceStatus defines the observed state of DBInstance
type DBInstanceStatus struct {
	Phase Phase `json:"phase"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// DBInstance is the Schema for the dbinstances API
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.phase`
type DBInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DBInstanceSpec   `json:"spec,omitempty"`
	Status DBInstanceStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// DBInstanceList contains a list of DBInstance
type DBInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []DBInstance `json:"items"`
}

func init() {
	SchemeBuilder.Register(&DBInstance{}, &DBInstanceList{})
}

func (in *DBInstance) GetDBInstanceID() string {
	out := fmt.Sprintf("%s-%s", in.GetNamespace(), in.GetName())
	if in.Spec.DBInstanceIdentifierOverride != "" {
		out = in.Spec.DBInstanceIdentifierOverride
	}
	return out
}
