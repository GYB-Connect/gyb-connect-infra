package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type DynamoDBStackProps struct {
	awscdk.StackProps
	Environment string
}

type DynamoDBStack struct {
	awscdk.Stack
	UserLogsTable awsdynamodb.Table
}

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

	var removalPolicy awscdk.RemovalPolicy
	if props.Environment == DEV_ENV {
		removalPolicy = awscdk.RemovalPolicy_DESTROY // Use DESTROY for development
	} else {
		removalPolicy = awscdk.RemovalPolicy_RETAIN // Use RETAIN for production
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

		// Add encryption at rest
		Encryption: awsdynamodb.TableEncryption_AWS_MANAGED,
		
		// Add time-to-live attribute for automatic cleanup
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
