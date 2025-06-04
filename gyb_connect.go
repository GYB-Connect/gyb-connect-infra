package main

import (
	"gyb_connect/stacks"
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
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
	isProduction := environment == stacks.PROD_ENV

	// 1. Create VPC Stack (only for production)
	var vpcStack *stacks.VpcStack
	if isProduction {
		vpcStack = stacks.NewVpcStack(app, "GybConnect-VpcStack", &stacks.VpcStackProps{
			StackProps: awscdk.StackProps{
				Env:         env,
				Description: jsii.String("VPC and networking infrastructure for GYB Connect"),
			},
		})
	}

	// 2. Create S3 Stack (independent)
	s3Stack := stacks.NewS3Stack(app, "GybConnect-S3Stack", &stacks.S3StackProps{
		Environment: environment,
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("S3 storage infrastructure for GYB Connect"),
		},
	})

	// 3. Create DynamoDB Stack (independent)
	dynamodbStack := stacks.NewDynamoDBStack(app, "GybConnect-DynamoDBStack", &stacks.DynamoDBStackProps{
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("DynamoDB database infrastructure for GYB Connect"),
		},
		Environment: environment,
	})

	// 4. Create RDS Stack (conditionally depends on VPC)
	var rdsVpc awsec2.IVpc
	if vpcStack != nil {
		rdsVpc = vpcStack.Vpc
	}

	rdsStack := stacks.NewRDSStack(app, "GybConnect-RDSStack", &stacks.RDSStackProps{
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("RDS PostgreSQL database infrastructure for GYB Connect"),
		},
		Vpc:         rdsVpc,
		Environment: environment,
	})

	// 5. Create API Gateway Stack (independent, but can reference other stacks later)
	apiStack := stacks.NewApiGatewayStack(app, "GybConnect-ApiGatewayStack", &stacks.ApiGatewayStackProps{
		StackProps: awscdk.StackProps{
			Env:         env,
			Description: jsii.String("API Gateway infrastructure for GYB Connect"),
		},
		Environment: environment,
	})

	// Add dependencies to ensure proper deployment order (only if VPC stack exists)
	if vpcStack != nil {
		rdsStack.AddDependency(vpcStack.Stack, jsii.String("VPC must be created before RDS"))
	}

	// Suppress unused variable warnings for now
	_ = s3Stack
	_ = dynamodbStack
	_ = apiStack

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
