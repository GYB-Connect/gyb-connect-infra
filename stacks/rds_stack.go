package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type RDSStackProps struct {
	awscdk.StackProps
	Vpc         awsec2.IVpc
	Environment string
}

type RDSStack struct {
	awscdk.Stack
	DatabaseInstance awsrds.DatabaseInstance
}

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
	// Production uses custom VPC (when both VPC is provided and environment is production)
	if props != nil && props.Vpc != nil && (props.Environment == PROD_ENV) {
		//* Production: Use provided VPC
		vpc = props.Vpc
		
		// Create security group for RDS in custom VPC
		dbSecurityGroup = awsec2.NewSecurityGroup(stack, jsii.String("DatabaseSecurityGroup"), &awsec2.SecurityGroupProps{
			Vpc:               vpc,
			Description:       jsii.String("Security group for RDS PostgreSQL database"),
			AllowAllOutbound: jsii.Bool(false),
		})

		// Allow inbound connections on PostgreSQL port from VPC
		dbSecurityGroup.AddIngressRule(
			awsec2.Peer_Ipv4(vpc.VpcCidrBlock()),
			awsec2.Port_Tcp(jsii.Number(5432)),
			jsii.String("Allow PostgreSQL access from VPC"),
			jsii.Bool(false),
		)

		// Create subnet group for RDS in custom VPC
		subnetGroup = awsrds.NewSubnetGroup(stack, jsii.String("DatabaseSubnetGroup"), &awsrds.SubnetGroupProps{
			Description: jsii.String("Subnet group for RDS PostgreSQL database"),
			Vpc:         vpc,
			VpcSubnets: &awsec2.SubnetSelection{
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
			},
		})
	} else {
		// Development: Use default VPC
		vpc = awsec2.Vpc_FromLookup(stack, jsii.String("DefaultVpc"), &awsec2.VpcLookupOptions{
			IsDefault: jsii.Bool(true),
		})
		
		// Create security group for RDS in default VPC
		dbSecurityGroup = awsec2.NewSecurityGroup(stack, jsii.String("DatabaseSecurityGroup"), &awsec2.SecurityGroupProps{
			Vpc:               vpc,
			Description:       jsii.String("Security group for RDS PostgreSQL database (development)"),
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
		// Note: Even for dev, consider using private subnets if available for better security
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
	// RDS PostgreSQL database
	rdsInstance := awsrds.NewDatabaseInstance(stack, jsii.String("GybConnectDatabase"), &awsrds.DatabaseInstanceProps{
		Engine: awsrds.DatabaseInstanceEngine_Postgres(&awsrds.PostgresInstanceEngineProps{
			Version: awsrds.PostgresEngineVersion_VER_15_13(),
		}),
		InstanceType:     awsec2.InstanceType_Of(awsec2.InstanceClass_BURSTABLE3, awsec2.InstanceSize_MICRO),
		DatabaseName:     jsii.String(databaseName),
		Credentials:      awsrds.Credentials_FromGeneratedSecret(jsii.String("gybconnect_admin"), nil),
		Vpc:              vpc,
		SubnetGroup:      subnetGroup,
		SecurityGroups: &[]awsec2.ISecurityGroup{
			dbSecurityGroup,
		},
		
		// Backup and maintenance settings
		AllowMajorVersionUpgrade: jsii.Bool(false),
		AutoMinorVersionUpgrade:  jsii.Bool(true),
		BackupRetention:          awscdk.Duration_Days(jsii.Number(7)),
		DeletionProtection:       jsii.Bool(false), // Set to true for production
		RemovalPolicy:            awscdk.RemovalPolicy_DESTROY, // Use RETAIN for production
		
		// Storage settings
		AllocatedStorage: jsii.Number(20),
		MaxAllocatedStorage: jsii.Number(100),
		StorageEncrypted: jsii.Bool(true),
		
		// Performance settings
		EnablePerformanceInsights: jsii.Bool(true),
		PerformanceInsightRetention: awsrds.PerformanceInsightRetention_DEFAULT,
		
		// Monitoring
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
