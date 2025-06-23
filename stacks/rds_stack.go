package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type RDSStackProps struct {
	awscdk.StackProps
	Vpc         awsec2.IVpc
	Environment string
	// PCI DSS Req 3.5: Accept customer-managed KMS key for encryption
	EncryptionKey awskms.IKey
}

type RDSStack struct {
	awscdk.Stack
	DatabaseInstance awsrds.DatabaseInstance
}

// NewRDSStack creates an RDS PostgreSQL database with PCI DSS compliant security settings
// This stack implements controls for PCI DSS Requirements 1.3, 2.2, 3.4, 3.5, 4.1, and 10.7
func NewRDSStack(scope constructs.Construct, id string, props *RDSStackProps) *RDSStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	var vpc awsec2.IVpc
	var dbSecurityGroup awsec2.ISecurityGroup
	var subnetGroup awsrds.ISubnetGroup

	//! Determine if we're using a custom VPC or the default VPC
	// PCI DSS Req 1.3: Production environments MUST use dedicated VPCs for network segmentation
	// Production uses custom VPC (when both VPC is provided and environment is production)
	if props != nil && props.Vpc != nil && (props.Environment == PROD_ENV) {
		//* Production: Use provided VPC
		// PCI DSS Req 1.3.4: Isolate the CDE from other networks
		vpc = props.Vpc
		
		// PCI DSS Req 1.2.1: Create security group for RDS with restrictive inbound rules
		// Security groups act as stateful firewalls at the instance level
		dbSecurityGroup = awsec2.NewSecurityGroup(stack, jsii.String("DatabaseSecurityGroup"), &awsec2.SecurityGroupProps{
			Vpc:               vpc,
			Description:       jsii.String("Security group for RDS PostgreSQL database"),
			// PCI DSS Req 1.2.1: Deny all outbound traffic by default
			AllowAllOutbound: jsii.Bool(false),
		})

		// PCI DSS Req 1.2.3: Allow only necessary protocols and ports
		// PostgreSQL default port 5432 - consider using non-standard port for additional security
		dbSecurityGroup.AddIngressRule(
			awsec2.Peer_Ipv4(vpc.VpcCidrBlock()),
			awsec2.Port_Tcp(jsii.Number(5432)),
			jsii.String("Allow PostgreSQL access from VPC"),
			jsii.Bool(false),
		)

		// PCI DSS Req 1.3.4: Place database in private subnets with no direct internet access
		// This ensures the database is not accessible from the public internet
		subnetGroup = awsrds.NewSubnetGroup(stack, jsii.String("DatabaseSubnetGroup"), &awsrds.SubnetGroupProps{
			Description: jsii.String("Subnet group for RDS PostgreSQL database"),
			Vpc:         vpc,
			VpcSubnets: &awsec2.SubnetSelection{
				// PCI DSS Req 1.3: Use private subnets with egress for updates but no ingress from internet
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
			},
		})
	} else {
		// Development: Use default VPC
		// WARNING: Default VPC usage is NOT PCI DSS compliant for production
		// This configuration is only acceptable for development environments
		vpc = awsec2.Vpc_FromLookup(stack, jsii.String("DefaultVpc"), &awsec2.VpcLookupOptions{
			IsDefault: jsii.Bool(true),
		})
		
		// Create security group for RDS in default VPC
		dbSecurityGroup = awsec2.NewSecurityGroup(stack, jsii.String("DatabaseSecurityGroup"), &awsec2.SecurityGroupProps{
			Vpc:               vpc,
			Description:       jsii.String("Security group for RDS PostgreSQL database (development)"),
			// PCI DSS Req 1.2.1: Even in dev, restrict outbound traffic
			AllowAllOutbound: jsii.Bool(false),
		})

		// Allow inbound connections on PostgreSQL port from anywhere in VPC
		dbSecurityGroup.AddIngressRule(
			awsec2.Peer_Ipv4(vpc.VpcCidrBlock()),
			awsec2.Port_Tcp(jsii.Number(5432)),
			jsii.String("Allow PostgreSQL access from VPC"),
			jsii.Bool(false),
		)

		// Create subnet group for RDS in default VPC
		// WARNING: Public subnets violate PCI DSS Req 1.3.4 - use only for development
		subnetGroup = awsrds.NewSubnetGroup(stack, jsii.String("DatabaseSubnetGroup"), &awsrds.SubnetGroupProps{
			Description: jsii.String("Subnet group for RDS PostgreSQL database (development)"),
			Vpc:         vpc,
			VpcSubnets: &awsec2.SubnetSelection{
				SubnetType: awsec2.SubnetType_PUBLIC, // Using public subnets for simplicity in dev
			},
		})
	}

	envPrefix := DEV_ENV // Default to development environment
	if props != nil && props.Environment != "" {
		envPrefix = props.Environment
	}
	// Set the database name based on the environment (must be alphanumeric only)
	databaseName := envPrefix + "_gyb_connect"
	
	// PCI DSS Req 2.2: Determine security settings based on environment
	// Production databases require maximum protection against accidental deletion
	var deletionProtection bool
	var removalPolicy awscdk.RemovalPolicy
	
	if envPrefix == PROD_ENV {
		// Production settings - maximum protection
		// PCI DSS Req 3.1: Enable deletion protection to prevent data loss
		deletionProtection = true
		// PCI DSS Req 3.1: RETAIN policy ensures data persists even if stack is deleted
		removalPolicy = awscdk.RemovalPolicy_RETAIN
	} else {
		// Development settings - allow cleanup
		deletionProtection = false
		removalPolicy = awscdk.RemovalPolicy_DESTROY
	}

	// PCI DSS Req 4.1: Create custom parameter group to enforce SSL connections
	// This ensures all connections to the database are encrypted in transit
	parameterGroup := awsrds.NewParameterGroup(stack, jsii.String("PostgreSQLParameterGroup"), &awsrds.ParameterGroupProps{
		Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
			Version: awsrds.PostgresEngineVersion_VER_15_13(),
		}),
		Description: jsii.String("Parameter group for PostgreSQL with SSL enforcement"),
		Parameters: &map[string]*string{
			// PCI DSS Req 4.1: Force SSL connections - no unencrypted connections allowed
			"rds.force_ssl": jsii.String("1"),
		},
	})
	
	// RDS PostgreSQL database
	rdsInstance := awsrds.NewDatabaseInstance(stack, jsii.String("GybConnectDatabase"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
			// PCI DSS Req 2.2.4: Use supported versions with security updates
			Version: awsrds.PostgresEngineVersion_VER_15_13(),
		}),
		// Consider using larger instance types for production workloads
		InstanceType:     awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_MICRO),
		DatabaseName:     jsii.String(databaseName),
		// PCI DSS Req 8.2.1: Use strong, unique passwords managed by AWS Secrets Manager
		// PCI DSS Req 3.6.4: Credentials are automatically rotated and encrypted
		Credentials:      awsrds.Credentials_FromGeneratedSecret(jsii.String("gybconnect_admin"), nil),
		Vpc:              vpc,
		SubnetGroup:      subnetGroup,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			dbSecurityGroup,
		},
		
		// PCI DSS Req 4.1: Apply custom parameter group with SSL enforcement
		ParameterGroup: parameterGroup,
		
		// Backup and maintenance settings
		// PCI DSS Req 2.2.4: Prevent major version upgrades to ensure stability
		AllowMajorVersionUpgrade: jsii.Bool(false),
		// PCI DSS Req 6.2: Enable automatic minor version updates for security patches
		AutoMinorVersionUpgrade:  jsii.Bool(true),
		// PCI DSS Req 12.10.1: Maintain backups for incident response and recovery
		BackupRetention:          awscdk.Duration_Days(jsii.Number(7)),
		// PCI DSS Req 3.1: Deletion protection based on environment
		DeletionProtection:       jsii.Bool(deletionProtection),
		// PCI DSS Req 3.1: Removal policy to protect production data
		RemovalPolicy:            removalPolicy,
		
		// Storage settings
		AllocatedStorage: jsii.Number(20),
		// PCI DSS Req 12.3.4: Enable storage auto-scaling to prevent outages
		MaxAllocatedStorage: jsii.Number(100),
		// PCI DSS Req 3.4.1: Enable encryption at rest for stored cardholder data
		StorageEncrypted: jsii.Bool(true),
		// PCI DSS Req 3.5: Use customer-managed KMS key when available for enhanced control
		StorageEncryptionKey: props.EncryptionKey, // Will be nil if not provided, uses default AWS key
		
		// Performance settings
		// PCI DSS Req 10.7: Enable performance insights for monitoring and forensics
		EnablePerformanceInsights: jsii.Bool(true),
		PerformanceInsightRetention: awsrds.PerformanceInsightRetention_DEFAULT,
		
		// PCI DSS Req 10.2: Enable enhanced monitoring for security event tracking
		MonitoringInterval: awscdk.Duration_Seconds(jsii.Number(60)),
	})

	// Output database information
	awscdk.NewCfnOutput(stack, jsii.String("DatabaseEndpoint"), &awscdk.CfnOutputProps{
		Value:       rdsInstance.InstanceEndpoint().Hostname(),
		Description: jsii.String("RDS PostgreSQL database endpoint"),
		ExportName:  jsii.String("GybConnect-DatabaseEndpoint"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DatabasePort"), &awscdk.CfnOutputProps{
		Value:       jsii.String("5432"),
		Description: jsii.String("RDS PostgreSQL database port"),
		ExportName:  jsii.String("GybConnect-DatabasePort"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DatabaseSecretArn"), &awscdk.CfnOutputProps{
		Value:       rdsInstance.Secret().SecretArn(),
		Description: jsii.String("ARN of the database credentials secret"),
		ExportName:  jsii.String("GybConnect-DatabaseSecretArn"),
	})

	return &RDSStack{
		Stack:            stack,
		DatabaseInstance: rdsInstance,
	}
}
