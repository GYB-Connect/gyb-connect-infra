package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type KmsStackProps struct {
	awscdk.StackProps
	Environment string
}

type KmsStack struct {
	awscdk.Stack
	S3Key       awskms.Key
	RDSKey      awskms.Key
	DynamoDBKey awskms.Key
	MacieKey    awskms.Key
	LoggingKey  awskms.Key
}

// NewKmsStack creates Customer-Managed Keys (CMKs) for PCI DSS compliant encryption
// This stack implements controls for PCI DSS Requirements 3.5 and 3.6
func NewKmsStack(scope constructs.Construct, id string, props *KmsStackProps) *KmsStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	envPrefix := DEV_ENV // Default to development environment
	if props != nil && props.Environment != "" {
		envPrefix = props.Environment
	}

	// PCI DSS Req 3.5: Create dedicated CMK for S3 encryption
	// This key will be used for all S3 bucket encryption in the environment
	s3Key := awskms.NewKey(stack, jsii.String("S3EncryptionKey"), &awskms.KeyProps{
		Description: jsii.String("Customer-managed key for S3 bucket encryption in " + envPrefix + " environment"),
		// PCI DSS Req 3.6.4: Enable automatic key rotation annually
		EnableKeyRotation: jsii.Bool(true),
		// PCI DSS Req 3.5.2: Define key usage policy with least privilege
		Policy: createS3KeyPolicy(envPrefix),
		// Environment-specific key alias for easier identification
		Alias: jsii.String(envPrefix + "/gyb-connect/s3"),
	})

	// PCI DSS Req 3.5: Create dedicated CMK for RDS encryption
	// This key will be used for RDS database encryption at rest
	rdsKey := awskms.NewKey(stack, jsii.String("RDSEncryptionKey"), &awskms.KeyProps{
		Description: jsii.String("Customer-managed key for RDS database encryption in " + envPrefix + " environment"),
		// PCI DSS Req 3.6.4: Enable automatic key rotation annually
		EnableKeyRotation: jsii.Bool(true),
		// PCI DSS Req 3.5.2: Define key usage policy with least privilege
		Policy: createRDSKeyPolicy(envPrefix),
		// Environment-specific key alias for easier identification
		Alias: jsii.String(envPrefix + "/gyb-connect/rds"),
	})

	// PCI DSS Req 3.5: Create dedicated CMK for DynamoDB encryption
	// This key will be used for DynamoDB table encryption at rest
	dynamoDBKey := awskms.NewKey(stack, jsii.String("DynamoDBEncryptionKey"), &awskms.KeyProps{
		Description: jsii.String("Customer-managed key for DynamoDB table encryption in " + envPrefix + " environment"),
		// PCI DSS Req 3.6.4: Enable automatic key rotation annually
		EnableKeyRotation: jsii.Bool(true),
		// PCI DSS Req 3.5.2: Define key usage policy with least privilege
		Policy: createDynamoDBKeyPolicy(envPrefix),
		// Environment-specific key alias for easier identification
		Alias: jsii.String(envPrefix + "/gyb-connect/dynamodb"),
	})

	// PCI DSS Req 3.5: Create dedicated CMK for Macie encryption
	// This key will be used for Macie data discovery and classification encryption
	macieKey := awskms.NewKey(stack, jsii.String("MacieEncryptionKey"), &awskms.KeyProps{
		Description: jsii.String("Customer-managed key for Macie data discovery and classification in " + envPrefix + " environment"),
		// PCI DSS Req 3.6.4: Enable automatic key rotation annually
		EnableKeyRotation: jsii.Bool(true),
		// PCI DSS Req 3.5.2: Define key usage policy with least privilege
		Policy: createMacieKeyPolicy(envPrefix),
		// Environment-specific key alias for easier identification
		Alias: jsii.String(envPrefix + "/gyb-connect/macie"),
	})

	// PCI DSS Req 10.5: Create dedicated CMK for logging encryption
	// This key will be used for CloudTrail, CloudWatch Logs, and SNS encryption
	loggingKey := awskms.NewKey(stack, jsii.String("LoggingEncryptionKey"), &awskms.KeyProps{
		Description: jsii.String("Customer-managed key for logging and monitoring encryption in " + envPrefix + " environment"),
		// PCI DSS Req 3.6.4: Enable automatic key rotation annually
		EnableKeyRotation: jsii.Bool(true),
		// PCI DSS Req 3.5.2: Define key usage policy with least privilege
		Policy: createLoggingKeyPolicy(envPrefix),
		// Environment-specific key alias for easier identification
		Alias: jsii.String(envPrefix + "/gyb-connect/logging"),
	})

	// Output key information for reference by other stacks
	awscdk.NewCfnOutput(stack, jsii.String("S3KeyId"), &awscdk.CfnOutputProps{
		Value:       s3Key.KeyId(),
		Description: jsii.String("ID of the S3 encryption key"),
		ExportName:  jsii.String("GybConnect-S3KeyId-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("S3KeyArn"), &awscdk.CfnOutputProps{
		Value:       s3Key.KeyArn(),
		Description: jsii.String("ARN of the S3 encryption key"),
		ExportName:  jsii.String("GybConnect-S3KeyArn-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("RDSKeyId"), &awscdk.CfnOutputProps{
		Value:       rdsKey.KeyId(),
		Description: jsii.String("ID of the RDS encryption key"),
		ExportName:  jsii.String("GybConnect-RDSKeyId-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("RDSKeyArn"), &awscdk.CfnOutputProps{
		Value:       rdsKey.KeyArn(),
		Description: jsii.String("ARN of the RDS encryption key"),
		ExportName:  jsii.String("GybConnect-RDSKeyArn-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DynamoDBKeyId"), &awscdk.CfnOutputProps{
		Value:       dynamoDBKey.KeyId(),
		Description: jsii.String("ID of the DynamoDB encryption key"),
		ExportName:  jsii.String("GybConnect-DynamoDBKeyId-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DynamoDBKeyArn"), &awscdk.CfnOutputProps{
		Value:       dynamoDBKey.KeyArn(),
		Description: jsii.String("ARN of the DynamoDB encryption key"),
		ExportName:  jsii.String("GybConnect-DynamoDBKeyArn-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("MacieKeyId"), &awscdk.CfnOutputProps{
		Value:       macieKey.KeyId(),
		Description: jsii.String("ID of the Macie encryption key"),
		ExportName:  jsii.String("GybConnect-MacieKeyId-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("MacieKeyArn"), &awscdk.CfnOutputProps{
		Value:       macieKey.KeyArn(),
		Description: jsii.String("ARN of the Macie encryption key"),
		ExportName:  jsii.String("GybConnect-MacieKeyArn-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("LoggingKeyId"), &awscdk.CfnOutputProps{
		Value:       loggingKey.KeyId(),
		Description: jsii.String("ID of the logging encryption key"),
		ExportName:  jsii.String("GybConnect-LoggingKeyId-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("LoggingKeyArn"), &awscdk.CfnOutputProps{
		Value:       loggingKey.KeyArn(),
		Description: jsii.String("ARN of the logging encryption key"),
		ExportName:  jsii.String("GybConnect-LoggingKeyArn-" + envPrefix),
	})

	return &KmsStack{
		Stack:       stack,
		S3Key:       s3Key,
		RDSKey:      rdsKey,
		DynamoDBKey: dynamoDBKey,
		MacieKey:    macieKey,
		LoggingKey:  loggingKey,
	}
}

// createS3KeyPolicy creates a restrictive KMS key policy for S3 encryption
// PCI DSS Req 3.5.2: Only allow specific IAM roles to use the key
func createS3KeyPolicy(environment string) awsiam.PolicyDocument {
	return awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		Statements: &[]awsiam.PolicyStatement{
			// Allow account root to manage the key
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Enable IAM User Permissions"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewAccountRootPrincipal(),
				},
				Actions: &[]*string{
					jsii.String("kms:*"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// Allow S3 service to use the key for server-side encryption
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Allow S3 Service"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewServicePrincipal(jsii.String("s3.amazonaws.com"), nil),
				},
				Actions: &[]*string{
					jsii.String("kms:Decrypt"),
					jsii.String("kms:GenerateDataKey"),
					jsii.String("kms:ReEncrypt*"),
					jsii.String("kms:DescribeKey"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// TODO: Add specific application role permissions when Lambda/AppRunner roles are created
			// This ensures only authorized services can encrypt/decrypt data
		},
	})
}

// createRDSKeyPolicy creates a restrictive KMS key policy for RDS encryption
// PCI DSS Req 3.5.2: Only allow RDS service and authorized roles to use the key
func createRDSKeyPolicy(environment string) awsiam.PolicyDocument {
	return awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		Statements: &[]awsiam.PolicyStatement{
			// Allow account root to manage the key
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Enable IAM User Permissions"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewAccountRootPrincipal(),
				},
				Actions: &[]*string{
					jsii.String("kms:*"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// Allow RDS service to use the key for encryption
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Allow RDS Service"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewServicePrincipal(jsii.String("rds.amazonaws.com"), nil),
				},
				Actions: &[]*string{
					jsii.String("kms:Decrypt"),
					jsii.String("kms:GenerateDataKey"),
					jsii.String("kms:ReEncrypt*"),
					jsii.String("kms:DescribeKey"),
					jsii.String("kms:CreateGrant"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// TODO: Add specific application role permissions when Lambda/AppRunner roles are created
		},
	})
}

// createDynamoDBKeyPolicy creates a restrictive KMS key policy for DynamoDB encryption
// PCI DSS Req 3.5.2: Only allow DynamoDB service and authorized roles to use the key
func createDynamoDBKeyPolicy(environment string) awsiam.PolicyDocument {
	return awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		Statements: &[]awsiam.PolicyStatement{
			// Allow account root to manage the key
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Enable IAM User Permissions"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewAccountRootPrincipal(),
				},
				Actions: &[]*string{
					jsii.String("kms:*"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// Allow DynamoDB service to use the key for encryption
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Allow DynamoDB Service"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewServicePrincipal(jsii.String("dynamodb.amazonaws.com"), nil),
				},
				Actions: &[]*string{
					jsii.String("kms:Decrypt"),
					jsii.String("kms:GenerateDataKey"),
					jsii.String("kms:ReEncrypt*"),
					jsii.String("kms:DescribeKey"),
					jsii.String("kms:CreateGrant"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// TODO: Add specific application role permissions when Lambda/AppRunner roles are created
		},
	})
}

// createMacieKeyPolicy creates a restrictive KMS key policy for Macie encryption
// PCI DSS Req 3.5.2: Only allow Macie service and authorized roles to use the key
func createMacieKeyPolicy(environment string) awsiam.PolicyDocument {
	return awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		Statements: &[]awsiam.PolicyStatement{
			// Allow account root to manage the key
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Enable IAM User Permissions"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewAccountRootPrincipal(),
				},
				Actions: &[]*string{
					jsii.String("kms:*"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// Allow Macie service to use the key for encryption
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Allow Macie Service"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewServicePrincipal(jsii.String("macie.amazonaws.com"), nil),
				},
				Actions: &[]*string{
					jsii.String("kms:Decrypt"),
					jsii.String("kms:GenerateDataKey"),
					jsii.String("kms:ReEncrypt*"),
					jsii.String("kms:DescribeKey"),
					jsii.String("kms:CreateGrant"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// TODO: Add specific application role permissions when Lambda/AppRunner roles are created
		},
	})
}

// createLoggingKeyPolicy creates a restrictive KMS key policy for logging encryption
// PCI DSS Req 10.5.2: Only allow logging services and authorized roles to use the key
func createLoggingKeyPolicy(environment string) awsiam.PolicyDocument {
	return awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		Statements: &[]awsiam.PolicyStatement{
			// Allow account root to manage the key
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Enable IAM User Permissions"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewAccountRootPrincipal(),
				},
				Actions: &[]*string{
					jsii.String("kms:*"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// Allow CloudTrail service to use the key
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Allow CloudTrail Service"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewServicePrincipal(jsii.String("cloudtrail.amazonaws.com"), nil),
				},
				Actions: &[]*string{
					jsii.String("kms:Decrypt"),
					jsii.String("kms:GenerateDataKey"),
					jsii.String("kms:ReEncrypt*"),
					jsii.String("kms:DescribeKey"),
					jsii.String("kms:CreateGrant"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// Allow CloudWatch Logs service to use the key
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Allow CloudWatch Logs Service"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewServicePrincipal(jsii.String("logs.amazonaws.com"), nil),
				},
				Actions: &[]*string{
					jsii.String("kms:Decrypt"),
					jsii.String("kms:GenerateDataKey"),
					jsii.String("kms:ReEncrypt*"),
					jsii.String("kms:DescribeKey"),
					jsii.String("kms:CreateGrant"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
			// Allow SNS service to use the key
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Sid:    jsii.String("Allow SNS Service"),
				Effect: awsiam.Effect_ALLOW,
				Principals: &[]awsiam.IPrincipal{
					awsiam.NewServicePrincipal(jsii.String("sns.amazonaws.com"), nil),
				},
				Actions: &[]*string{
					jsii.String("kms:Decrypt"),
					jsii.String("kms:GenerateDataKey"),
					jsii.String("kms:ReEncrypt*"),
					jsii.String("kms:DescribeKey"),
					jsii.String("kms:CreateGrant"),
				},
				Resources: &[]*string{
					jsii.String("*"),
				},
			}),
		},
	})
} 