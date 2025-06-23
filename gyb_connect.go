package main

import (
	"gyb_connect/stacks"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/jsii-runtime-go"
)

type GybConnectStackProps struct {
	awscdk.StackProps
}

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Define common environment
	env := env()

	// Get environment from context or environment variable
	environment := getEnvironment(app)

	// 1. Create KMS Stack first (PCI DSS Req 3.5)
	// This must be created before other stacks that depend on the encryption keys
	kmsStack := stacks.NewKmsStack(app, "GybConnect-KmsStack", &stacks.KmsStackProps{
		Environment: environment,
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("KMS encryption keys for PCI DSS compliance"),
		},
	})

	// 2. Create Security Stack (PCI DSS Req 5.2, 6.2, 11.2.3, 11.3)
	// Enable GuardDuty, Inspector, and Security Hub for threat detection and vulnerability management
	securityStack := stacks.NewSecurityStack(app, "GybConnect-SecurityStack", &stacks.SecurityStackProps{
		Environment: environment,
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("Security services for threat detection and vulnerability management"),
		},
	})

	// 3. Create VPC Stack
	vpcStack := stacks.NewVpcStack(app, "GybConnect-VpcStack", &stacks.VpcStackProps{
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("VPC and networking infrastructure for GYB Connect"),
		},
	})

	// 4. Create S3 Stack with customer-managed encryption key
	s3Stack := stacks.NewS3Stack(app, "GybConnect-S3Stack", &stacks.S3StackProps{
		Environment: environment,
		// PCI DSS Req 3.5: Pass customer-managed KMS key for S3 encryption
		EncryptionKey: kmsStack.S3Key,
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("S3 storage infrastructure for GYB Connect with customer-managed encryption"),
		},
	})

	// 5. Create DynamoDB Stack with customer-managed encryption key
	dynamodbStack := stacks.NewDynamoDBStack(app, "GybConnect-DynamoDBStack", &stacks.DynamoDBStackProps{
		Environment: environment,
		// PCI DSS Req 3.5: Pass customer-managed KMS key for DynamoDB encryption
		EncryptionKey: kmsStack.DynamoDBKey,
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("DynamoDB database infrastructure for GYB Connect with customer-managed encryption"),
		},
	})

	// 6. Create IAM Stack (PCI DSS Req 7.1, 7.2, 8.1)
	// Least privilege roles and MFA enforcement for all access
	iamStack := stacks.NewIAMStack(app, "GybConnect-IAMStack", &stacks.IAMStackProps{
		Environment: environment,
		S3Bucket: s3Stack.UploadsBucket,
		DynamoDBTable: dynamodbStack.UserLogsTable,
		S3KmsKey: kmsStack.S3Key,
		DynamoDBKmsKey: kmsStack.DynamoDBKey,
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("IAM roles and policies for least privilege access control and MFA enforcement"),
		},
	})

	// 7. Create Logging Stack (PCI DSS Req 10)
	// Centralized logging with CloudTrail, CloudWatch, and real-time alerting
	loggingStack := stacks.NewLoggingStack(app, "GybConnect-LoggingStack", &stacks.LoggingStackProps{
		Environment: environment,
		LoggingKmsKey: kmsStack.LoggingKey,
		SecurityAlertEmail: "security@gybconnect.com", // Replace with your security team email
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("Centralized logging and monitoring infrastructure for PCI DSS Requirement 10"),
		},
	})

	// 8. Create RDS Stack with customer-managed encryption key (depends on VPC)
	rdsStack := stacks.NewRDSStack(app, "GybConnect-RDSStack", &stacks.RDSStackProps{
		Vpc:         vpcStack.Vpc,
		Environment: environment,
		// PCI DSS Req 3.5: Pass customer-managed KMS key for RDS encryption
		EncryptionKey: kmsStack.RDSKey,
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("RDS PostgreSQL database infrastructure for GYB Connect with customer-managed encryption and SSL enforcement"),
		},
	})

	// 9. Create API Gateway Stack with custom domain and certificate
	// PCI DSS Req 4.1: Configure custom domain with TLS 1.2+ enforcement
	var certificate awscertificatemanager.ICertificate
	var domainName string
	
	// Get certificate ARN from environment variable
	certArn := os.Getenv("ACM_CERTIFICATE_ARN")
	if certArn != "" {
		certificate = awscertificatemanager.Certificate_FromCertificateArn(
			app,
			jsii.String("ApiCertificate"),
			jsii.String(certArn),
		)
		
		// Set domain name based on environment
		if environment == stacks.PROD_ENV {
			domainName = "api.gybconnect.com" // Replace with your actual production domain
		} else {
			domainName = "api-dev.gybconnect.com" // Replace with your actual dev domain
		}
	}

	apiStack := stacks.NewApiGatewayStack(app, "GybConnect-ApiGatewayStack", &stacks.ApiGatewayStackProps{
		Environment: environment,
		// PCI DSS Req 4.1: Configure custom domain and certificate for TLS 1.2+ enforcement
		DomainName:  domainName,
		Certificate: certificate,
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("API Gateway infrastructure for GYB Connect with PCI DSS security controls and custom domain"),
		},
	})

	// Add dependencies to ensure proper deployment order
	// KMS stack must be deployed first
	s3Stack.AddDependency(kmsStack.Stack, jsii.String("KMS keys must be created before S3 encryption"))
	dynamodbStack.AddDependency(kmsStack.Stack, jsii.String("KMS keys must be created before DynamoDB encryption"))
	iamStack.AddDependency(s3Stack.Stack, jsii.String("S3 bucket must be created before IAM policies"))
	iamStack.AddDependency(dynamodbStack.Stack, jsii.String("DynamoDB table must be created before IAM policies"))
	iamStack.AddDependency(kmsStack.Stack, jsii.String("KMS keys must be created before IAM policies"))
	loggingStack.AddDependency(kmsStack.Stack, jsii.String("KMS keys must be created before logging encryption"))
	rdsStack.AddDependency(kmsStack.Stack, jsii.String("KMS keys must be created before RDS encryption"))
	rdsStack.AddDependency(vpcStack.Stack, jsii.String("VPC must be created before RDS"))

	// Suppress unused variable warnings for now
	_ = apiStack
	_ = securityStack
	_ = iamStack
	_ = loggingStack

	app.Synth(nil)
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// Use the AWS Account from CLI configuration and preferred region
	// This is required for VPC lookups and other context-dependent features
	region := os.Getenv("CDK_DEFAULT_REGION")
	if region == "" {
		region = "us-west-1" // Default to us-west-1 for GYB Connect
	}

	return &awscdk.Environment{
		Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
		Region:  jsii.String(region),
	}
}

// getEnvironment determines the deployment environment from CDK context or environment variables
func getEnvironment(app awscdk.App) string {
	// Check CDK context first
	if env := app.Node().TryGetContext(jsii.String("environment")); env != nil {
		if envStr, ok := env.(string); ok {
			// Normalize environment values
			if envStr == "production" {
				return "prod"
			}
			return envStr
		}
	}

	// Fall back to environment variable
	if env := os.Getenv("DEPLOY_ENV"); env != "" {
		// Normalize environment values
		if env == "production" {
			return "prod"
		}
		return env
	}

	// Default to development
	return stacks.DEV_ENV
}
