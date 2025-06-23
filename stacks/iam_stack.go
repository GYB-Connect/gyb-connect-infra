package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type IAMStackProps struct {
	awscdk.StackProps
	Environment string
	// Pass in resources that need IAM access
	S3Bucket      awss3.IBucket
	DynamoDBTable awsdynamodb.ITable
	S3KmsKey      awskms.IKey
	DynamoDBKmsKey awskms.IKey
}

type IAMStack struct {
	awscdk.Stack
	ApiLambdaRole         awsiam.Role
	DataProcessingRole    awsiam.Role
	ReadOnlyRole         awsiam.Role
	ComplianceAuditorRole awsiam.Role
}

// NewIAMStack creates IAM roles and policies for PCI DSS compliant access control
// This stack implements controls for PCI DSS Requirements 7.1, 7.2, and 8.1
func NewIAMStack(scope constructs.Construct, id string, props *IAMStackProps) *IAMStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	envPrefix := DEV_ENV // Default to development environment
	if props != nil && props.Environment != "" {
		envPrefix = props.Environment
	}

	// PCI DSS Req 8.1: Create a boundary policy that enforces MFA for all actions
	// This will be attached to all roles to ensure MFA is required
	mfaBoundaryPolicy := awsiam.NewManagedPolicy(stack, jsii.String("MFABoundaryPolicy"), &awsiam.ManagedPolicyProps{
		ManagedPolicyName: jsii.String(envPrefix + "-mfa-boundary-policy"),
		Description:       jsii.String("PCI DSS Req 8.4: Enforces MFA for all privileged actions"),
		Statements: &[]awsiam.PolicyStatement{
			// Allow basic account info without MFA
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("iam:ListAccountAliases"),
					jsii.String("iam:ListUsers"),
					jsii.String("iam:GetAccountPasswordPolicy"),
					jsii.String("iam:GetAccountSummary"),
				},
				Resources: &[]*string{jsii.String("*")},
			}),
			// Allow MFA self-service
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_ALLOW,
				Actions: &[]*string{
					jsii.String("iam:CreateVirtualMFADevice"),
					jsii.String("iam:DeleteVirtualMFADevice"),
					jsii.String("iam:EnableMFADevice"),
					jsii.String("iam:ResyncMFADevice"),
					jsii.String("iam:ListMFADevices"),
				},
				Resources: &[]*string{
					jsii.String("arn:aws:iam::*:mfa/${aws:username}"),
					jsii.String("arn:aws:iam::*:user/${aws:username}"),
				},
			}),
			// Deny all other actions without MFA
			awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
				Effect: awsiam.Effect_DENY,
				NotActions: &[]*string{
					jsii.String("iam:CreateVirtualMFADevice"),
					jsii.String("iam:EnableMFADevice"),
					jsii.String("iam:GetUser"),
					jsii.String("iam:ListMFADevices"),
					jsii.String("iam:ResyncMFADevice"),
					jsii.String("sts:GetSessionToken"),
				},
				Resources: &[]*string{jsii.String("*")},
				Conditions: &map[string]interface{}{
					"BoolIfExists": map[string]interface{}{
						"aws:MultiFactorAuthPresent": "false",
					},
				},
			}),
		},
	})

	// PCI DSS Req 7.2: Create least privilege role for API Lambda functions
	apiLambdaRole := awsiam.NewRole(stack, jsii.String("ApiLambdaExecutionRole"), &awsiam.RoleProps{
		RoleName:    jsii.String(envPrefix + "-gyb-api-lambda-role"),
		Description: jsii.String("PCI DSS Req 7.2: Least privilege role for API Lambda functions"),
		AssumedBy: awsiam.NewCompositePrincipal(
			awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
		),
		// PCI DSS Req 7.1: Apply permission boundary to enforce MFA
		PermissionsBoundary: mfaBoundaryPolicy,
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")),
		},
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"ApiLambdaPolicy": createApiLambdaPolicy(props),
		},
	})

	// Add tags to the role
	awscdk.Tags_Of(apiLambdaRole).Add(jsii.String("Environment"), jsii.String(envPrefix), nil)
	awscdk.Tags_Of(apiLambdaRole).Add(jsii.String("PCI-DSS-Requirement"), jsii.String("7.2"), nil)
	awscdk.Tags_Of(apiLambdaRole).Add(jsii.String("Purpose"), jsii.String("API-Lambda-Execution"), nil)

	// PCI DSS Req 7.2: Create least privilege role for data processing tasks
	dataProcessingRole := awsiam.NewRole(stack, jsii.String("DataProcessingRole"), &awsiam.RoleProps{
		RoleName:    jsii.String(envPrefix + "-gyb-data-processing-role"),
		Description: jsii.String("PCI DSS Req 7.2: Least privilege role for data processing tasks"),
		AssumedBy: awsiam.NewCompositePrincipal(
			awsiam.NewServicePrincipal(jsii.String("lambda.amazonaws.com"), nil),
			awsiam.NewServicePrincipal(jsii.String("ecs-tasks.amazonaws.com"), nil),
		),
		PermissionsBoundary: mfaBoundaryPolicy,
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("service-role/AWSLambdaBasicExecutionRole")),
		},
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"DataProcessingPolicy": createDataProcessingPolicy(props),
		},
	})

	// Add tags to the role
	awscdk.Tags_Of(dataProcessingRole).Add(jsii.String("Environment"), jsii.String(envPrefix), nil)
	awscdk.Tags_Of(dataProcessingRole).Add(jsii.String("PCI-DSS-Requirement"), jsii.String("7.2"), nil)
	awscdk.Tags_Of(dataProcessingRole).Add(jsii.String("Purpose"), jsii.String("Data-Processing"), nil)

	// PCI DSS Req 7.2.4: Create read-only role for monitoring and audit
	readOnlyRole := awsiam.NewRole(stack, jsii.String("ReadOnlyAccessRole"), &awsiam.RoleProps{
		RoleName:    jsii.String(envPrefix + "-gyb-readonly-role"),
		Description: jsii.String("PCI DSS Req 7.2.4: Read-only role for monitoring and audit"),
		AssumedBy: awsiam.NewCompositePrincipal(
			// Allow specific users from IAM Identity Center
			awsiam.NewFederatedPrincipal(
				jsii.String("arn:aws:iam::*:saml-provider/IdentityCenter"),
				&map[string]interface{}{
					"StringEquals": map[string]interface{}{
						"SAML:aud": "https://signin.aws.amazon.com/saml",
					},
				},
				jsii.String("sts:AssumeRoleWithSAML"),
			),
		),
		PermissionsBoundary: mfaBoundaryPolicy,
		ManagedPolicies: &[]awsiam.IManagedPolicy{
			awsiam.ManagedPolicy_FromAwsManagedPolicyName(jsii.String("ReadOnlyAccess")),
		},
		// PCI DSS Req 8.1.3: Maximum session duration of 15 minutes for privileged access
		MaxSessionDuration: awscdk.Duration_Minutes(jsii.Number(15)),
	})

	// Add tags to the role
	awscdk.Tags_Of(readOnlyRole).Add(jsii.String("Environment"), jsii.String(envPrefix), nil)
	awscdk.Tags_Of(readOnlyRole).Add(jsii.String("PCI-DSS-Requirement"), jsii.String("7.2.4"), nil)
	awscdk.Tags_Of(readOnlyRole).Add(jsii.String("Purpose"), jsii.String("Read-Only-Access"), nil)

	// PCI DSS Req 7.2.5: Create compliance auditor role with specific permissions
	complianceAuditorRole := awsiam.NewRole(stack, jsii.String("ComplianceAuditorRole"), &awsiam.RoleProps{
		RoleName:    jsii.String(envPrefix + "-gyb-compliance-auditor-role"),
		Description: jsii.String("PCI DSS Req 7.2.5: Role for compliance auditors with limited access"),
		AssumedBy: awsiam.NewFederatedPrincipal(
			jsii.String("arn:aws:iam::*:saml-provider/IdentityCenter"),
			&map[string]interface{}{
				"StringEquals": map[string]interface{}{
					"SAML:aud": "https://signin.aws.amazon.com/saml",
				},
			},
			jsii.String("sts:AssumeRoleWithSAML"),
		),
		PermissionsBoundary: mfaBoundaryPolicy,
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"ComplianceAuditPolicy": createComplianceAuditorPolicy(),
		},
		MaxSessionDuration: awscdk.Duration_Hours(jsii.Number(1)),
	})

	// Add tags to the role
	awscdk.Tags_Of(complianceAuditorRole).Add(jsii.String("Environment"), jsii.String(envPrefix), nil)
	awscdk.Tags_Of(complianceAuditorRole).Add(jsii.String("PCI-DSS-Requirement"), jsii.String("7.2.5"), nil)
	awscdk.Tags_Of(complianceAuditorRole).Add(jsii.String("Purpose"), jsii.String("Compliance-Audit"), nil)

	// PCI DSS Req 7.2.2: Create IAM Access Analyzer for continuous monitoring
	_ = awscdk.NewCfnResource(stack, jsii.String("AccessAnalyzer"), &awscdk.CfnResourceProps{
		Type: jsii.String("AWS::AccessAnalyzer::Analyzer"),
		Properties: &map[string]interface{}{
			"Type": "ACCOUNT",
			"AnalyzerName": envPrefix + "-gyb-access-analyzer",
			"Tags": []map[string]string{
				{
					"Key":   "Environment",
					"Value": envPrefix,
				},
				{
					"Key":   "PCI-DSS-Requirement",
					"Value": "7.2.2",
				},
			},
		},
	})

	// Output important information
	awscdk.NewCfnOutput(stack, jsii.String("ApiLambdaRoleArn"), &awscdk.CfnOutputProps{
		Value:       apiLambdaRole.RoleArn(),
		Description: jsii.String("ARN of the API Lambda execution role"),
		ExportName:  jsii.String("GybConnect-ApiLambdaRoleArn-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DataProcessingRoleArn"), &awscdk.CfnOutputProps{
		Value:       dataProcessingRole.RoleArn(),
		Description: jsii.String("ARN of the data processing role"),
		ExportName:  jsii.String("GybConnect-DataProcessingRoleArn-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ReadOnlyRoleArn"), &awscdk.CfnOutputProps{
		Value:       readOnlyRole.RoleArn(),
		Description: jsii.String("ARN of the read-only access role"),
		ExportName:  jsii.String("GybConnect-ReadOnlyRoleArn-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ComplianceAuditorRoleArn"), &awscdk.CfnOutputProps{
		Value:       complianceAuditorRole.RoleArn(),
		Description: jsii.String("ARN of the compliance auditor role"),
		ExportName:  jsii.String("GybConnect-ComplianceAuditorRoleArn-" + envPrefix),
	})

	return &IAMStack{
		Stack:                 stack,
		ApiLambdaRole:         apiLambdaRole,
		DataProcessingRole:    dataProcessingRole,
		ReadOnlyRole:         readOnlyRole,
		ComplianceAuditorRole: complianceAuditorRole,
	}
}

// createApiLambdaPolicy creates a least privilege policy for API Lambda functions
func createApiLambdaPolicy(props *IAMStackProps) awsiam.PolicyDocument {
	statements := []awsiam.PolicyStatement{
		// PCI DSS Req 7.2: Only allow specific S3 operations on the uploads bucket
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("s3:GetObject"),
				jsii.String("s3:PutObject"),
				jsii.String("s3:DeleteObject"),
			},
			Resources: &[]*string{
				jsii.String(*props.S3Bucket.BucketArn() + "/*"),
			},
		}),
		// Allow listing bucket contents
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("s3:ListBucket"),
			},
			Resources: &[]*string{
				props.S3Bucket.BucketArn(),
			},
		}),
		// PCI DSS Req 7.2: Only allow specific DynamoDB operations
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("dynamodb:GetItem"),
				jsii.String("dynamodb:PutItem"),
				jsii.String("dynamodb:Query"),
				jsii.String("dynamodb:UpdateItem"),
			},
			Resources: &[]*string{
				props.DynamoDBTable.TableArn(),
				jsii.String(*props.DynamoDBTable.TableArn() + "/index/*"),
			},
		}),
		// PCI DSS Req 3.5: Allow use of KMS keys for encryption/decryption
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("kms:Decrypt"),
				jsii.String("kms:GenerateDataKey"),
			},
			Resources: &[]*string{
				props.S3KmsKey.KeyArn(),
				props.DynamoDBKmsKey.KeyArn(),
			},
		}),
	}

	return awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		Statements: &statements,
	})
}

// createDataProcessingPolicy creates a policy for batch data processing
func createDataProcessingPolicy(props *IAMStackProps) awsiam.PolicyDocument {
	statements := []awsiam.PolicyStatement{
		// Read-only access to S3 bucket
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("s3:GetObject"),
				jsii.String("s3:ListBucket"),
			},
			Resources: &[]*string{
				props.S3Bucket.BucketArn(),
				jsii.String(*props.S3Bucket.BucketArn() + "/*"),
			},
		}),
		// Read/write access to DynamoDB for processing results
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("dynamodb:BatchGetItem"),
				jsii.String("dynamodb:BatchWriteItem"),
				jsii.String("dynamodb:Query"),
				jsii.String("dynamodb:Scan"),
			},
			Resources: &[]*string{
				props.DynamoDBTable.TableArn(),
			},
		}),
		// KMS permissions for data encryption
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("kms:Decrypt"),
				jsii.String("kms:GenerateDataKey"),
			},
			Resources: &[]*string{
				props.S3KmsKey.KeyArn(),
				props.DynamoDBKmsKey.KeyArn(),
			},
		}),
	}

	return awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		Statements: &statements,
	})
}

// createComplianceAuditorPolicy creates a policy for compliance auditors
func createComplianceAuditorPolicy() awsiam.PolicyDocument {
	statements := []awsiam.PolicyStatement{
		// Read-only access to CloudTrail logs
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("cloudtrail:LookupEvents"),
				jsii.String("cloudtrail:GetTrailStatus"),
				jsii.String("cloudtrail:DescribeTrails"),
			},
			Resources: &[]*string{jsii.String("*")},
		}),
		// Read Security Hub findings
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("securityhub:GetFindings"),
				jsii.String("securityhub:GetComplianceSummary"),
				jsii.String("securityhub:GetEnabledStandards"),
			},
			Resources: &[]*string{jsii.String("*")},
		}),
		// Read GuardDuty findings
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("guardduty:GetFindings"),
				jsii.String("guardduty:ListFindings"),
				jsii.String("guardduty:GetDetector"),
			},
			Resources: &[]*string{jsii.String("*")},
		}),
		// Read KMS key policies (but not use the keys)
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("kms:DescribeKey"),
				jsii.String("kms:GetKeyPolicy"),
				jsii.String("kms:ListAliases"),
			},
			Resources: &[]*string{jsii.String("*")},
		}),
		// Read IAM policies and roles
		awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
			Effect: awsiam.Effect_ALLOW,
			Actions: &[]*string{
				jsii.String("iam:GetRole"),
				jsii.String("iam:GetRolePolicy"),
				jsii.String("iam:ListRolePolicies"),
				jsii.String("iam:ListAttachedRolePolicies"),
				jsii.String("iam:GetPolicy"),
				jsii.String("iam:GetPolicyVersion"),
			},
			Resources: &[]*string{jsii.String("*")},
		}),
	}

	return awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
		Statements: &statements,
	})
} 