package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscloudtrail"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskms"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type LoggingStackProps struct {
	awscdk.StackProps
	Environment string
	// PCI DSS Req 10.5: Accept KMS key for log encryption
	LoggingKmsKey awskms.IKey
	// Email for security alerts
	SecurityAlertEmail string
}

type LoggingStack struct {
	awscdk.Stack
	CloudTrail          awscloudtrail.Trail
	LoggingBucket       awss3.Bucket
	SecurityAlertsTopic awssns.Topic
	LogGroup            awslogs.LogGroup
}

// NewLoggingStack creates centralized logging infrastructure for PCI DSS Requirement 10
// This stack implements controls for PCI DSS Requirements 10.1, 10.2, 10.3, 10.4, 10.5, 10.6, and 10.7
func NewLoggingStack(scope constructs.Construct, id string, props *LoggingStackProps) *LoggingStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	envPrefix := DEV_ENV // Default to development environment
	if props != nil && props.Environment != "" {
		envPrefix = props.Environment
	}

	// PCI DSS Req 10.3.1: Create centralized logging S3 bucket with Object Lock
	// This ensures logs cannot be tampered with or deleted
	loggingBucket := awss3.NewBucket(stack, jsii.String("CentralLoggingBucket"), &awss3.BucketProps{
		BucketName: jsii.String(envPrefix + "-gyb-central-logs"),
		// PCI DSS Req 10.4.1: Enable versioning to protect against accidental deletion
		Versioned: jsii.Bool(true),
		// PCI DSS Req 10.4.2: Block all public access to prevent log exposure
		PublicReadAccess:  jsii.Bool(false),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		
		// PCI DSS Req 10.5.1: Enable Object Lock for immutable logs
		ObjectLockEnabled: jsii.Bool(true),
		
		// PCI DSS Req 10.5.2: Encrypt logs at rest using customer-managed key
		Encryption: awss3.BucketEncryption_KMS,
		EncryptionKey: props.LoggingKmsKey,
		
		// Production logs should never be deleted accidentally
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
		AutoDeleteObjects: jsii.Bool(false),
		
		// PCI DSS Req 10.7.2: Lifecycle rules for cost optimization after retention period
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:     jsii.String("archive-old-logs"),
				Enabled: jsii.Bool(true),
				// Archive logs older than 90 days to Glacier for cost optimization
				Transitions: &[]*awss3.Transition{
					{
						StorageClass: awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(90)),
					},
					{
						StorageClass: awss3.StorageClass_DEEP_ARCHIVE(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)),
					},
				},
			},
		},
	})

	// PCI DSS Req 10.6.2: Create SNS topic for security alerts
	securityAlertsTopic := awssns.NewTopic(stack, jsii.String("SecurityAlertsTopic"), &awssns.TopicProps{
		TopicName:   jsii.String(envPrefix + "-gyb-security-alerts"),
		DisplayName: jsii.String("GYB Connect Security Alerts"),
		// PCI DSS Req 10.5.2: Encrypt SNS messages
		MasterKey: props.LoggingKmsKey,
	})

	// Subscribe security team email to alerts if provided
	if props != nil && props.SecurityAlertEmail != "" {
		awssns.NewSubscription(stack, jsii.String("SecurityAlertsEmail"), &awssns.SubscriptionProps{
			Topic:    securityAlertsTopic,
			Endpoint: jsii.String(props.SecurityAlertEmail),
			Protocol: awssns.SubscriptionProtocol_EMAIL,
		})
	}

	// PCI DSS Req 10.2.1: Create CloudWatch Log Group for application logs
	logGroup := awslogs.NewLogGroup(stack, jsii.String("ApplicationLogGroup"), &awslogs.LogGroupProps{
		LogGroupName: jsii.String("/gyb-connect/" + envPrefix + "/application"),
		// PCI DSS Req 10.7.1: Retain logs for minimum 1 year
		Retention: awslogs.RetentionDays_THIRTEEN_MONTHS,
		// PCI DSS Req 10.5.2: Encrypt logs using customer-managed key
		EncryptionKey: props.LoggingKmsKey,
		// Prevent accidental deletion
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
	})

	// PCI DSS Req 10.2.2: Enable CloudTrail for comprehensive API logging
	cloudTrail := awscloudtrail.NewTrail(stack, jsii.String("OrganizationCloudTrail"), &awscloudtrail.TrailProps{
		TrailName: jsii.String(envPrefix + "-gyb-cloudtrail"),
		// PCI DSS Req 10.2.1: Log all API calls across all regions
		IsMultiRegionTrail: jsii.Bool(true),
		// PCI DSS Req 10.2.3: Include global service events (IAM, STS, etc.)
		IncludeGlobalServiceEvents: jsii.Bool(true),
		// PCI DSS Req 10.2.4: Enable log file validation to detect tampering
		EnableFileValidation: jsii.Bool(true),
		
		// Store CloudTrail logs in the central logging bucket
		Bucket: loggingBucket,
		S3KeyPrefix: jsii.String("cloudtrail-logs/"),
		
		// PCI DSS Req 10.2.5: Send to CloudWatch for real-time monitoring
		SendToCloudWatchLogs: jsii.Bool(true),
		CloudWatchLogGroup: logGroup,
		CloudWatchLogsRetention: awslogs.RetentionDays_THIRTEEN_MONTHS,
	})

	// PCI DSS Req 10.6.3: Create EventBridge rules for real-time security events
	
	// Rule for IAM policy changes
	iamPolicyChangeRule := awsevents.NewRule(stack, jsii.String("IAMPolicyChangeRule"), &awsevents.RuleProps{
		RuleName:    jsii.String(envPrefix + "-iam-policy-changes"),
		Description: jsii.String("PCI DSS Req 10.6: Detect IAM policy changes"),
		EventPattern: &awsevents.EventPattern{
			Source:      &[]*string{jsii.String("aws.iam")},
			DetailType:  &[]*string{jsii.String("AWS API Call via CloudTrail")},
			Detail: &map[string]interface{}{
				"eventName": []string{
					"AttachUserPolicy",
					"DetachUserPolicy",
					"AttachRolePolicy",
					"DetachRolePolicy",
					"CreateRole",
					"DeleteRole",
					"PutUserPolicy",
					"PutRolePolicy",
					"DeleteUserPolicy",
					"DeleteRolePolicy",
				},
			},
		},
	})
	iamPolicyChangeRule.AddTarget(awseventstargets.NewSnsTopic(securityAlertsTopic, &awseventstargets.SnsTopicProps{
		Message: awsevents.RuleTargetInput_FromText(jsii.String("PCI DSS Alert: IAM policy change detected. Event: <eventName> by user: <userIdentity.userName> at <eventTime>")),
	}))

	// Rule for S3 bucket policy changes
	s3PolicyChangeRule := awsevents.NewRule(stack, jsii.String("S3PolicyChangeRule"), &awsevents.RuleProps{
		RuleName:    jsii.String(envPrefix + "-s3-policy-changes"),
		Description: jsii.String("PCI DSS Req 10.6: Detect S3 bucket policy changes"),
		EventPattern: &awsevents.EventPattern{
			Source:     &[]*string{jsii.String("aws.s3")},
			DetailType: &[]*string{jsii.String("AWS API Call via CloudTrail")},
			Detail: &map[string]interface{}{
				"eventName": []string{
					"PutBucketPolicy",
					"DeleteBucketPolicy",
					"PutBucketAcl",
					"PutBucketPublicAccessBlock",
					"DeleteBucketPublicAccessBlock",
				},
			},
		},
	})
	s3PolicyChangeRule.AddTarget(awseventstargets.NewSnsTopic(securityAlertsTopic, &awseventstargets.SnsTopicProps{
		Message: awsevents.RuleTargetInput_FromText(jsii.String("PCI DSS Alert: S3 bucket policy change detected. Bucket: <requestParameters.bucketName> Event: <eventName>")),
	}))

	// Rule for KMS key policy changes
	kmsKeyChangeRule := awsevents.NewRule(stack, jsii.String("KMSKeyChangeRule"), &awsevents.RuleProps{
		RuleName:    jsii.String(envPrefix + "-kms-key-changes"),
		Description: jsii.String("PCI DSS Req 10.6: Detect KMS key policy changes"),
		EventPattern: &awsevents.EventPattern{
			Source:     &[]*string{jsii.String("aws.kms")},
			DetailType: &[]*string{jsii.String("AWS API Call via CloudTrail")},
			Detail: &map[string]interface{}{
				"eventName": []string{
					"CreateKey",
					"DisableKey",
					"ScheduleKeyDeletion",
					"PutKeyPolicy",
					"CreateAlias",
					"DeleteAlias",
				},
			},
		},
	})
	kmsKeyChangeRule.AddTarget(awseventstargets.NewSnsTopic(securityAlertsTopic, &awseventstargets.SnsTopicProps{
		Message: awsevents.RuleTargetInput_FromText(jsii.String("PCI DSS Alert: KMS key change detected. Key: <requestParameters.keyId> Event: <eventName>")),
	}))

	// PCI DSS Req 10.2.6: Create metric filters for CloudWatch Logs
	// These turn log events into CloudWatch metrics for alerting
	
	// Root account usage metric filter
	awslogs.NewMetricFilter(stack, jsii.String("RootAccountUsageMetricFilter"), &awslogs.MetricFilterProps{
		LogGroup:    logGroup,
		FilterName:  jsii.String("RootAccountUsage"),
		FilterPattern: awslogs.FilterPattern_Literal(jsii.String("{ ($.userIdentity.type = \"Root\") && ($.userIdentity.invokedBy NOT EXISTS) && ($.eventType != \"AwsServiceEvent\") }")),
		MetricNamespace: jsii.String("CloudTrailMetrics"),
		MetricName:      jsii.String("RootAccountUsageCount"),
		MetricValue:     jsii.String("1"),
		DefaultValue:    jsii.Number(0),
	})

	// Failed console login metric filter
	awslogs.NewMetricFilter(stack, jsii.String("ConsoleLoginFailureMetricFilter"), &awslogs.MetricFilterProps{
		LogGroup:    logGroup,
		FilterName:  jsii.String("ConsoleLoginFailures"),
		FilterPattern: awslogs.FilterPattern_Literal(jsii.String("{ ($.eventName = ConsoleLogin) && ($.responseElements.ConsoleLogin = \"Failure\") }")),
		MetricNamespace: jsii.String("CloudTrailMetrics"),
		MetricName:      jsii.String("ConsoleLoginFailures"),
		MetricValue:     jsii.String("1"),
		DefaultValue:    jsii.Number(0),
	})

	// CloudTrail configuration changes metric filter
	awslogs.NewMetricFilter(stack, jsii.String("CloudTrailChangesMetricFilter"), &awslogs.MetricFilterProps{
		LogGroup:    logGroup,
		FilterName:  jsii.String("CloudTrailChanges"),
		FilterPattern: awslogs.FilterPattern_Literal(jsii.String("{ ($.eventName = CreateTrail) || ($.eventName = UpdateTrail) || ($.eventName = DeleteTrail) || ($.eventName = StartLogging) || ($.eventName = StopLogging) }")),
		MetricNamespace: jsii.String("CloudTrailMetrics"),
		MetricName:      jsii.String("CloudTrailChanges"),
		MetricValue:     jsii.String("1"),
		DefaultValue:    jsii.Number(0),
	})

	// Add tags to all resources
	awscdk.Tags_Of(stack).Add(jsii.String("Environment"), jsii.String(envPrefix), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("PCI-DSS-Requirement"), jsii.String("10"), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("Purpose"), jsii.String("Centralized-Logging"), nil)

	// Output important information
	awscdk.NewCfnOutput(stack, jsii.String("LoggingBucketName"), &awscdk.CfnOutputProps{
		Value:       loggingBucket.BucketName(),
		Description: jsii.String("Name of the central logging S3 bucket"),
		ExportName:  jsii.String("GybConnect-LoggingBucket-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("SecurityAlertsTopicArn"), &awscdk.CfnOutputProps{
		Value:       securityAlertsTopic.TopicArn(),
		Description: jsii.String("ARN of the security alerts SNS topic"),
		ExportName:  jsii.String("GybConnect-SecurityAlertsTopic-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("CloudTrailArn"), &awscdk.CfnOutputProps{
		Value:       cloudTrail.TrailArn(),
		Description: jsii.String("ARN of the CloudTrail for audit logging"),
		ExportName:  jsii.String("GybConnect-CloudTrail-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ApplicationLogGroupName"), &awscdk.CfnOutputProps{
		Value:       logGroup.LogGroupName(),
		Description: jsii.String("Name of the application log group"),
		ExportName:  jsii.String("GybConnect-ApplicationLogGroup-" + envPrefix),
	})

	return &LoggingStack{
		Stack:               stack,
		CloudTrail:          cloudTrail,
		LoggingBucket:       loggingBucket,
		SecurityAlertsTopic: securityAlertsTopic,
		LogGroup:            logGroup,
	}
} 