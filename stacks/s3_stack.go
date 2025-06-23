package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type S3StackProps struct {
	awscdk.StackProps
	Environment string
	// PCI DSS Req 3.5: Accept customer-managed KMS key for encryption
	EncryptionKey awskms.IKey
}

type S3Stack struct {
	awscdk.Stack
	UploadsBucket awss3.Bucket
}

// NewS3Stack creates an S3 bucket for file uploads with PCI DSS compliant security settings
// This stack implements controls for PCI DSS Requirements 2.2, 3.4, 3.5, and 4.1
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

	// PCI DSS Req 2.2: Determine security settings based on environment
	// Production environments require stricter controls to protect cardholder data
	var removalPolicy awscdk.RemovalPolicy
	var autoDeleteObjects bool
	var corsOrigins []*string

	if envPrefix == PROD_ENV {
		// Production settings - more restrictive
		// PCI DSS Req 2.2: RETAIN policy prevents accidental deletion of data that may contain cardholder information
		removalPolicy = awscdk.RemovalPolicy_RETAIN
		// PCI DSS Req 3.1: Disable auto-deletion to ensure data retention policies are followed
		autoDeleteObjects = false
		// PCI DSS Req 2.2: Restrict CORS to specific production domains only
		corsOrigins = []*string{
			jsii.String("https://app.gybconnect.com"), // Replace with your actual production domain
			jsii.String("https://www.gybconnect.com"),
		}
	} else {
		// Development settings - allows easier cleanup for testing
		removalPolicy = awscdk.RemovalPolicy_DESTROY
		autoDeleteObjects = true
		// Development origins - broader access for testing purposes
		corsOrigins = []*string{
			jsii.String("http://localhost:3000"),
			jsii.String("http://localhost:5173"), // Vite default
			jsii.String("https://dev.gybconnect.com"), // Replace with your actual dev domain
		}
	}

	// PCI DSS Req 3.5.1: Determine encryption configuration based on available KMS key
	var encryptionConfig awss3.BucketEncryption
	var encryptionKey awskms.IKey
	if props != nil && props.EncryptionKey != nil {
		// PCI DSS Req 3.5: Use customer-managed KMS key for enhanced control and auditability
		encryptionConfig = awss3.BucketEncryption_KMS
		encryptionKey = props.EncryptionKey
	} else {
		// Fallback to S3-managed encryption (not recommended for production)
		// PCI DSS Req 3.4.1: Still provides encryption at rest but with less control
		encryptionConfig = awss3.BucketEncryption_S3_MANAGED
	}

	// S3 bucket for file storage
	uploadsBucket := awss3.NewBucket(stack, jsii.String("GybUploadsS3"), &awss3.BucketProps{
		BucketName:        bucketName,
		Versioned:         jsii.Bool(true),
		// PCI DSS Req 7.1: Deny public access to prevent unauthorized data exposure
		PublicReadAccess:  jsii.Bool(false),
		// PCI DSS Req 7.1: Block all public access as an additional security layer
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		// PCI DSS Req 3.1: Removal policy based on environment to protect production data
		RemovalPolicy:     removalPolicy,
		// PCI DSS Req 3.1: Auto-deletion settings aligned with data retention requirements
		AutoDeleteObjects: jsii.Bool(autoDeleteObjects),
		
		// PCI DSS Req 3.5: Use customer-managed KMS key for encryption at rest
		// This provides better audit trails and key management control
		Encryption: encryptionConfig,
		EncryptionKey: encryptionKey,
		
		// PCI DSS Req 3.1: Lifecycle rules for cost optimization and data hygiene
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:     jsii.String("delete-incomplete-multipart-uploads"),
				Enabled: jsii.Bool(true),
				// Clean up incomplete uploads to prevent data leakage
				AbortIncompleteMultipartUploadAfter: awscdk.Duration_Days(jsii.Number(7)),
			},
		},
	})

	// PCI DSS Req 2.2: Configure CORS with principle of least privilege
	// Only allow specific origins rather than wildcards to prevent unauthorized access
	uploadsBucket.AddCorsRule(&awss3.CorsRule{
		// PCI DSS Req 6.4.2: Restrict origins to known, trusted sources only
		AllowedOrigins: &corsOrigins,
		// Define specific HTTP methods needed by the application
		AllowedMethods: &[]awss3.HttpMethods{
			awss3.HttpMethods_GET,
			awss3.HttpMethods_POST,
			awss3.HttpMethods_PUT,
			awss3.HttpMethods_DELETE,
		},
		// TODO: Consider restricting headers to specific required headers instead of wildcard
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
