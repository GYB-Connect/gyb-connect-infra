package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsevents"
	"github.com/aws/aws-cdk-go/awscdk/v2/awseventstargets"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsguardduty"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsinspectorv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecurityhub"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type SecurityStackProps struct {
	awscdk.StackProps
	Environment string
}

type SecurityStack struct {
	awscdk.Stack
	GuardDutyDetector awsguardduty.CfnDetector
	SecurityTopic     awssns.Topic
}

// NewSecurityStack creates security services for PCI DSS compliant threat detection and vulnerability management
// This stack implements controls for PCI DSS Requirements 5.2, 6.2, 11.2.3, and 11.3
func NewSecurityStack(scope constructs.Construct, id string, props *SecurityStackProps) *SecurityStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	envPrefix := DEV_ENV // Default to development environment
	if props != nil && props.Environment != "" {
		envPrefix = props.Environment
	}

	// PCI DSS Req 11.3: Create SNS topic for security alerts
	// This topic will receive all critical security findings
	securityTopic := awssns.NewTopic(stack, jsii.String("SecurityAlertsTopic"), &awssns.TopicProps{
		TopicName:   jsii.String(envPrefix + "-gyb-security-alerts"),
		DisplayName: jsii.String("GYB Connect Security Alerts"),
	})

	// Add email subscription for security alerts (update with your actual email)
	// TODO: Uncomment and replace with your security team's email address
	// securityTopic.AddSubscription(awssnssubscriptions.NewEmailSubscription(
	// 	jsii.String("security@gybconnect.com"),
	// 	nil,
	// ))

	// PCI DSS Req 5.2 & 11.3: Enable Amazon GuardDuty for malware protection and intrusion detection
	// GuardDuty acts as our IDS and provides continuous threat detection
	guarddutyDetector := awsguardduty.NewCfnDetector(stack, jsii.String("GuardDutyDetector"), &awsguardduty.CfnDetectorProps{
		Enable: jsii.Bool(true),
		// PCI DSS Req 5.2: Enable malware protection for EBS volumes
		DataSources: &awsguardduty.CfnDetector_CFNDataSourceConfigurationsProperty{
			MalwareProtection: &awsguardduty.CfnDetector_CFNMalwareProtectionConfigurationProperty{
				ScanEc2InstanceWithFindings: &awsguardduty.CfnDetector_CFNScanEc2InstanceWithFindingsConfigurationProperty{
					EbsVolumes: jsii.Bool(true),
				},
			},
			S3Logs: &awsguardduty.CfnDetector_CFNS3LogsConfigurationProperty{
				Enable: jsii.Bool(true),
			},
		},
		// PCI DSS Req 11.3: Configure finding publishing frequency for timely alerts
		FindingPublishingFrequency: jsii.String("FIFTEEN_MINUTES"),
		Tags: &[]*awsguardduty.CfnDetector_TagItemProperty{
			{
				Key:   jsii.String("Environment"),
				Value: jsii.String(envPrefix),
			},
			{
				Key:   jsii.String("PCI-DSS-Requirement"),
				Value: jsii.String("5.2,11.3"),
			},
		},
	})

	// PCI DSS Req 6.2 & 11.2.3: Enable Amazon Inspector for vulnerability scanning
	// Inspector continuously scans for software vulnerabilities and network exposure
	_ = awsinspectorv2.NewCfnFilter(stack, jsii.String("InspectorConfiguration"), &awsinspectorv2.CfnFilterProps{
		Name:        jsii.String(envPrefix + "-gyb-inspector-filter"),
		Description: jsii.String("PCI DSS vulnerability scanning configuration"),
		FilterAction: jsii.String("SUPPRESS"),
		FilterCriteria: &awsinspectorv2.CfnFilter_FilterCriteriaProperty{
			// Suppress findings for development resources to reduce noise
			ResourceTags: &[]*awsinspectorv2.CfnFilter_MapFilterProperty{
				{
					Comparison: jsii.String("EQUALS"),
					Key:        jsii.String("Environment"),
					Value:      jsii.String("dev"),
				},
			},
		},
	})

	// PCI DSS Req 11.3: Enable AWS Security Hub for centralized security findings
	// Security Hub aggregates findings from GuardDuty, Inspector, and other services
	_ = awssecurityhub.NewCfnHub(stack, jsii.String("SecurityHub"), &awssecurityhub.CfnHubProps{
		// Enable compliance standards
		EnableDefaultStandards: jsii.Bool(true),
		Tags: &map[string]*string{
			"Environment":       jsii.String(envPrefix),
			"PCI-DSS-Requirement": jsii.String("11.3"),
		},
	})

	// PCI DSS Req 11.3: Enable PCI DSS standard in Security Hub
	_ = awssecurityhub.NewCfnStandard(stack, jsii.String("PCIDSSStandard"), &awssecurityhub.CfnStandardProps{
		StandardsArn: jsii.String("arn:aws:securityhub:::ruleset/pci-dss/v/3.2.1"),
		DisabledStandardsControls: &[]*awssecurityhub.CfnStandard_StandardsControlProperty{
			// Disable controls that don't apply to our serverless architecture
			{
				StandardsControlArn: jsii.String("arn:aws:securityhub:*:*:control/pci-dss/v/3.2.1/PCI.EC2.1"),
				Reason:             jsii.String("Not using EC2 instances"),
			},
		},
	})

	// PCI DSS Req 11.3 & 10.4: Create EventBridge rules for critical security findings
	// High severity findings require immediate attention
	_ = awsevents.NewRule(stack, jsii.String("HighSeverityFindingsRule"), &awsevents.RuleProps{
		RuleName:    jsii.String(envPrefix + "-high-severity-findings"),
		Description: jsii.String("Alert on high severity security findings from GuardDuty and Inspector"),
		EventPattern: &awsevents.EventPattern{
			Source: &[]*string{
				jsii.String("aws.guardduty"),
				jsii.String("aws.inspector2"),
				jsii.String("aws.securityhub"),
			},
			DetailType: &[]*string{
				jsii.String("GuardDuty Finding"),
				jsii.String("Inspector2 Finding"),
				jsii.String("Security Hub Findings - Imported"),
			},
			Detail: &map[string]interface{}{
				"severity": []interface{}{
					map[string]interface{}{
						"numeric": []interface{}{">", 7}, // High severity: 7-10
					},
				},
			},
		},
		Targets: &[]awsevents.IRuleTarget{
			awseventstargets.NewSnsTopic(securityTopic, nil),
		},
	})

	// PCI DSS Req 5.1.2: Alert on malware detection
	_ = awsevents.NewRule(stack, jsii.String("MalwareDetectionRule"), &awsevents.RuleProps{
		RuleName:    jsii.String(envPrefix + "-malware-detection"),
		Description: jsii.String("Alert on malware detection from GuardDuty"),
		EventPattern: &awsevents.EventPattern{
			Source: &[]*string{
				jsii.String("aws.guardduty"),
			},
			DetailType: &[]*string{
				jsii.String("GuardDuty Finding"),
			},
			Detail: &map[string]interface{}{
				"type": []interface{}{
					map[string]interface{}{
						"prefix": "Execution:EC2/MaliciousFile",
					},
				},
			},
		},
		Targets: &[]awsevents.IRuleTarget{
			awseventstargets.NewSnsTopic(securityTopic, nil),
		},
	})

	// PCI DSS Req 11.2.3: Alert on vulnerability findings
	_ = awsevents.NewRule(stack, jsii.String("VulnerabilityFindingsRule"), &awsevents.RuleProps{
		RuleName:    jsii.String(envPrefix + "-vulnerability-findings"),
		Description: jsii.String("Alert on critical vulnerabilities from Inspector"),
		EventPattern: &awsevents.EventPattern{
			Source: &[]*string{
				jsii.String("aws.inspector2"),
			},
			DetailType: &[]*string{
				jsii.String("Inspector2 Finding"),
			},
			Detail: &map[string]interface{}{
				"severity": []interface{}{
					"CRITICAL",
					"HIGH",
				},
			},
		},
		Targets: &[]awsevents.IRuleTarget{
			awseventstargets.NewSnsTopic(securityTopic, nil),
		},
	})

	// Output important information
	awscdk.NewCfnOutput(stack, jsii.String("SecurityTopicArn"), &awscdk.CfnOutputProps{
		Value:       securityTopic.TopicArn(),
		Description: jsii.String("ARN of the security alerts SNS topic"),
		ExportName:  jsii.String("GybConnect-SecurityTopicArn-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("GuardDutyDetectorId"), &awscdk.CfnOutputProps{
		Value:       guarddutyDetector.AttrId(),
		Description: jsii.String("ID of the GuardDuty detector"),
		ExportName:  jsii.String("GybConnect-GuardDutyDetectorId-" + envPrefix),
	})

	return &SecurityStack{
		Stack:             stack,
		GuardDutyDetector: guarddutyDetector,
		SecurityTopic:     securityTopic,
	}
} 