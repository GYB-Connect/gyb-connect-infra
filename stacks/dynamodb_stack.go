package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DynamoDBStackProps struct {
	awscdk.StackProps
	Environment string
	// PCI DSS Req 3.5: Accept customer-managed KMS key for encryption
	EncryptionKey awskms.IKey
}

type DynamoDBStack struct {
	awscdk.Stack
	UserLogsTable awsdynamodb.Table
}

// NewDynamoDBStack creates a DynamoDB table with PCI DSS compliant security settings
// This stack implements controls for PCI DSS Requirements 3.1, 3.4, and 3.5
func NewDynamoDBStack(scope constructs.Construct, id string, props *DynamoDBStackProps) *DynamoDBStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)
	envPrefix := DEV_ENV // Default to development environment
	if props != nil && props.Environment != "" {
		envPrefix = props.Environment
	}
	tableName := envPrefix + "-gyb-user-logs"

	// PCI DSS Req 3.1: Configure retention and deletion policies based on environment
	var removalPolicy awscdk.RemovalPolicy
	if props.Environment == DEV_ENV {
		removalPolicy = awscdk.RemovalPolicy_DESTROY // Use DESTROY for development
	} else {
		// PCI DSS Req 3.1: Use RETAIN for production to protect cardholder data
		removalPolicy = awscdk.RemovalPolicy_RETAIN // Use RETAIN for production
	}

	// PCI DSS Req 3.5.1: Determine encryption configuration based on available KMS key
	var encryptionConfig awsdynamodb.TableEncryption
	if props != nil && props.EncryptionKey != nil {
		// PCI DSS Req 3.5: Use customer-managed KMS key for enhanced control and auditability
		encryptionConfig = awsdynamodb.TableEncryption_CUSTOMER_MANAGED
	} else {
		// Fallback to AWS-managed encryption (not recommended for production)
		// PCI DSS Req 3.4.1: Still provides encryption at rest but with less control
		encryptionConfig = awsdynamodb.TableEncryption_AWS_MANAGED
	}

	// DynamoDB table for user logs
	userLogsTable := awsdynamodb.NewTable(stack, jsii.String("GybUserLogsTable"), &awsdynamodb.TableProps{
		TableName:     jsii.String(tableName),
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("userId"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("timestamp"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		BillingMode:   awsdynamodb.BillingMode_PAY_PER_REQUEST,
		RemovalPolicy: removalPolicy,

		// PCI DSS Req 3.5: Use customer-managed KMS key for encryption at rest when available
		// This provides better audit trails and key management control
		Encryption: encryptionConfig,
		EncryptionKey: props.EncryptionKey, // Will be nil if not provided, which is fine
		
		// PCI DSS Req 3.1: Add time-to-live attribute for automatic cleanup of old data
		// This helps with data retention compliance
		TimeToLiveAttribute: jsii.String("ttl"),
	})

	//TODO Add Global Secondary Index for querying by action type
	// userLogsTable.AddGlobalSecondaryIndex(&awsdynamodb.GlobalSecondaryIndexProps{
	// 	IndexName: jsii.String("ActionTypeIndex"),
	// 	PartitionKey: &awsdynamodb.Attribute{
	// 		Name: jsii.String("actionType"),
	// 		Type: awsdynamodb.AttributeType_STRING,
	// 	},
	// 	SortKey: &awsdynamodb.Attribute{
	// 		Name: jsii.String("timestamp"),
	// 		Type: awsdynamodb.AttributeType_STRING,
	// 	},
	// })

	// Output table information
	awscdk.NewCfnOutput(stack, jsii.String("DynamoDBTableName"), &awscdk.CfnOutputProps{
		Value:       userLogsTable.TableName(),
		Description: jsii.String("Name of the DynamoDB table for user logs"),
		ExportName:  jsii.String("GybConnect-DynamoDBTableName"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DynamoDBTableArn"), &awscdk.CfnOutputProps{
		Value:       userLogsTable.TableArn(),
		Description: jsii.String("ARN of the DynamoDB table for user logs"),
		ExportName:  jsii.String("GybConnect-DynamoDBTableArn"),
	})

	return &DynamoDBStack{
		Stack:         stack,
		UserLogsTable: userLogsTable,
	}
}
