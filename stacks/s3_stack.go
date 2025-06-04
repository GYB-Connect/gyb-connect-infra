package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type S3StackProps struct {
	awscdk.StackProps
	Environment string
}

type S3Stack struct {
	awscdk.Stack
	UploadsBucket awss3.Bucket
}

func NewS3Stack(scope constructs.Construct, id string, props *S3StackProps) *S3Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Use environment prefix for bucket name
	envPrefix := DEV_ENV // Default to development environment
	if props != nil && props.Environment != "" {
		envPrefix = props.Environment
	}
	bucketName := jsii.String(envPrefix + "-gyb-uploads")

	// S3 bucket for file storage
	uploadsBucket := awss3.NewBucket(stack, jsii.String("GybUploadsS3"), &awss3.BucketProps{
		BucketName:        bucketName,
		Versioned:         jsii.Bool(true),
		PublicReadAccess:  jsii.Bool(false),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY, // Use RETAIN for production
		AutoDeleteObjects: jsii.Bool(true), // Use false for production
		
		// Add server-side encryption
		Encryption: awss3.BucketEncryption_S3_MANAGED,
		
		// Add lifecycle rules for cost optimization
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:     jsii.String("delete-incomplete-multipart-uploads"),
				Enabled: jsii.Bool(true),
				AbortIncompleteMultipartUploadAfter: awscdk.Duration_Days(jsii.Number(7)),
			},
		},
	})

	// Add CORS configuration for web uploads
	uploadsBucket.AddCorsRule(&awss3.CorsRule{
		AllowedOrigins: &[]*string{jsii.String("*")}, // Restrict this in production
		AllowedMethods: &[]awss3.HttpMethods{
			awss3.HttpMethods_GET,
			awss3.HttpMethods_POST,
			awss3.HttpMethods_PUT,
			awss3.HttpMethods_DELETE,
		},
		AllowedHeaders: &[]*string{
			jsii.String("*"),
		},
		MaxAge: jsii.Number(3000),
	})

	// Output bucket information
	awscdk.NewCfnOutput(stack, jsii.String("S3BucketName"), &awscdk.CfnOutputProps{
		Value:       uploadsBucket.BucketName(),
		Description: jsii.String("Name of the S3 bucket for file uploads"),
		ExportName:  jsii.String("GybConnect-S3BucketName"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("S3BucketArn"), &awscdk.CfnOutputProps{
		Value:       uploadsBucket.BucketArn(),
		Description: jsii.String("ARN of the S3 bucket for file uploads"),
		ExportName:  jsii.String("GybConnect-S3BucketArn"),
	})

	return &S3Stack{
		Stack:         stack,
		UploadsBucket: uploadsBucket,
	}
}
