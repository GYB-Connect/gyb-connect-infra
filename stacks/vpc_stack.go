package stacks

import (
	"fmt"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslogs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type VpcStackProps struct {
	awscdk.StackProps
}

type VpcStack struct {
	awscdk.Stack
	Vpc awsec2.Vpc
}

func NewVpcStack(scope constructs.Construct, id string, props *VpcStackProps) *VpcStack {
	var stackProps awscdk.StackProps
	if props != nil {
		stackProps = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &stackProps)

	// VPC for shared infrastructure
	vpc := awsec2.NewVpc(stack, jsii.String("GybConnectVpc"), &awsec2.VpcProps{
		MaxAzs:             jsii.Number(2),
		VpcName:            jsii.String("gyb-connect-vpc"),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				// Public subnet for load balancers and NAT gateways
				Name:       jsii.String("public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				// Private subnet for application servers, AppRunner, Lambda etc.
				Name:       jsii.String("private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
			{
				// Isolated subnet for databases and other resources that don't need internet access
				Name:       jsii.String("data"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// Gateway Endpoint for S3 (allows private access to S3)
	vpc.AddGatewayEndpoint(jsii.String("S3GatewayEndpoint"), &awsec2.GatewayVpcEndpointOptions{
		Service: awsec2.GatewayVpcEndpointAwsService_S3(),
	})

	// Gateway Endpoint for DynamoDB
	vpc.AddGatewayEndpoint(jsii.String("DynamoDBGatewayEndpoint"), &awsec2.GatewayVpcEndpointOptions{
		Service: awsec2.GatewayVpcEndpointAwsService_DYNAMODB(),
	})

	// Interface Endpoint for KMS (critical for CMK usage mentioned in Req 3)
	vpc.AddInterfaceEndpoint(jsii.String("KmsInterfaceEndpoint"), &awsec2.InterfaceVpcEndpointOptions{
		Service: awsec2.InterfaceVpcEndpointAwsService_KMS(),
		// Place the endpoint in the isolated data subnets
		Subnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
	})

	// Create CloudWatch Log Group for VPC Flow Logs
	vpcFlowLogsGroup := awslogs.NewLogGroup(stack, jsii.String("VpcFlowLogsGroup"), &awslogs.LogGroupProps{
		LogGroupName:  jsii.String("/aws/vpc/flowlogs"),
		Retention:     awslogs.RetentionDays_ONE_YEAR, // PCI DSS requires log retention
		RemovalPolicy: awscdk.RemovalPolicy_RETAIN,    // Critical for compliance - logs must be preserved
	})

	// Create IAM role for VPC Flow Logs to write to CloudWatch
	flowLogsRole := awsiam.NewRole(stack, jsii.String("VpcFlowLogsRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("vpc-flow-logs.amazonaws.com"), nil),
		RoleName:  jsii.String("gyb-connect-vpc-flow-logs-role"),
		InlinePolicies: &map[string]awsiam.PolicyDocument{
			"FlowLogsDeliveryRolePolicy": awsiam.NewPolicyDocument(&awsiam.PolicyDocumentProps{
				Statements: &[]awsiam.PolicyStatement{
					awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
						Effect: awsiam.Effect_ALLOW,
						Actions: &[]*string{
							jsii.String("logs:CreateLogGroup"),
							jsii.String("logs:CreateLogStream"),
							jsii.String("logs:PutLogEvents"),
							jsii.String("logs:DescribeLogGroups"),
							jsii.String("logs:DescribeLogStreams"),
						},
						Resources: &[]*string{
							vpcFlowLogsGroup.LogGroupArn(),
						},
					}),
				},
			}),
		},
	})

	// Create S3 bucket for long-term VPC Flow Logs storage (centralized logging)
	// This bucket should ideally be in a dedicated Security Account for production
	vpcFlowLogsBucket := awss3.NewBucket(stack, jsii.String("VpcFlowLogsBucket"), &awss3.BucketProps{
		BucketName:        jsii.String("gyb-connect-vpc-flow-logs-" + *awscdk.Aws_ACCOUNT_ID() + "-" + *awscdk.Aws_REGION()),
		Versioned:         jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN, // Critical for compliance
		AutoDeleteObjects: jsii.Bool(false),            // Never auto-delete security logs
		PublicReadAccess:  jsii.Bool(false),
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		Encryption:        awss3.BucketEncryption_S3_MANAGED, // TODO: Use KMS in production
		LifecycleRules: &[]*awss3.LifecycleRule{
			{
				Id:      jsii.String("VpcFlowLogsRetention"),
				Enabled: jsii.Bool(true),
				// Transition to cheaper storage classes for long-term retention
				Transitions: &[]*awss3.Transition{
					{
						StorageClass:    awss3.StorageClass_INFREQUENT_ACCESS(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(30)),
					},
					{
						StorageClass:    awss3.StorageClass_GLACIER(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(90)),
					},
					{
						StorageClass:    awss3.StorageClass_DEEP_ARCHIVE(),
						TransitionAfter: awscdk.Duration_Days(jsii.Number(365)),
					},
				},
				// PCI DSS requires at least 1 year retention, keeping 7 years for forensics
				Expiration: awscdk.Duration_Days(jsii.Number(2555)), // 7 years
			},
		},
	})

	// Enable VPC Flow Logs to CloudWatch Logs
	awsec2.NewFlowLog(stack, jsii.String("VpcFlowLogsToCloudWatch"), &awsec2.FlowLogProps{
		ResourceType:           awsec2.FlowLogResourceType_FromVpc(vpc),
		TrafficType:            awsec2.FlowLogTrafficType_ALL, // Capture ALL traffic (accepted and rejected)
		Destination:            awsec2.FlowLogDestination_ToCloudWatchLogs(vpcFlowLogsGroup, flowLogsRole),
		FlowLogName:            jsii.String("gyb-connect-vpc-flow-logs-cw"),
		MaxAggregationInterval: awsec2.FlowLogMaxAggregationInterval_ONE_MINUTE, // More granular for security monitoring
	})

	// Enable VPC Flow Logs to S3 (for long-term storage and centralized logging)
	awsec2.NewFlowLog(stack, jsii.String("VpcFlowLogsToS3"), &awsec2.FlowLogProps{
		ResourceType:           awsec2.FlowLogResourceType_FromVpc(vpc),
		TrafficType:            awsec2.FlowLogTrafficType_ALL, // Capture ALL traffic
		Destination:            awsec2.FlowLogDestination_ToS3(vpcFlowLogsBucket, jsii.String("vpc-flow-logs/"), &awsec2.S3DestinationOptions{}),
		FlowLogName:            jsii.String("gyb-connect-vpc-flow-logs-s3"),
		MaxAggregationInterval: awsec2.FlowLogMaxAggregationInterval_ONE_MINUTE,
	})

	privateSubnets := vpc.PrivateSubnets()
	dataSubnets := vpc.IsolatedSubnets()
	publicSubnets := vpc.PublicSubnets()

	// Collect all CIDR blocks for use in NACL rules
	var privateCidrs []string
	var publicCidrs []string
	var dataCidrs []string

	// Populate CIDR arrays
	for _, subnet := range *privateSubnets {
		privateCidrs = append(privateCidrs, *subnet.Ipv4CidrBlock())
	}
	for _, subnet := range *publicSubnets {
		publicCidrs = append(publicCidrs, *subnet.Ipv4CidrBlock())
	}
	for _, subnet := range *dataSubnets {
		dataCidrs = append(dataCidrs, *subnet.Ipv4CidrBlock())
	}

	// NACL for Public Subnets (Load Balancers, NAT Gateways)
	publicNacl := awsec2.NewNetworkAcl(stack, jsii.String("PublicSubnetNacl"), &awsec2.NetworkAclProps{
		Vpc: vpc,
		SubnetSelection: &awsec2.SubnetSelection{
			Subnets: publicSubnets,
		},
		NetworkAclName: jsii.String("gyb-connect-public-nacl"),
	})

	// Public subnet NACL rules - Allow HTTP/HTTPS inbound from internet
	publicNacl.AddEntry(jsii.String("AllowHTTPSInbound"), &awsec2.CommonNetworkAclEntryOptions{
		Cidr:       awsec2.AclCidr_AnyIpv4(),
		RuleNumber: jsii.Number(100),
		Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(443)), // HTTPS
		Direction:  awsec2.TrafficDirection_INGRESS,
	})

	publicNacl.AddEntry(jsii.String("AllowHTTPInbound"), &awsec2.CommonNetworkAclEntryOptions{
		Cidr:       awsec2.AclCidr_AnyIpv4(),
		RuleNumber: jsii.Number(110),
		Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(80)), // HTTP (for redirects to HTTPS)
		Direction:  awsec2.TrafficDirection_INGRESS,
	})

	// Allow ephemeral ports for return traffic
	publicNacl.AddEntry(jsii.String("AllowEphemeralInbound"), &awsec2.CommonNetworkAclEntryOptions{
		Cidr:       awsec2.AclCidr_AnyIpv4(),
		RuleNumber: jsii.Number(120),
		Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1024), jsii.Number(65535)),
		Direction:  awsec2.TrafficDirection_INGRESS,
	})

	//* Allow outbound traffic to PRIVATE subnets - loop through all private subnets
	if len(*privateSubnets) > 0 {
		for i, subnet := range *privateSubnets {
			privateCidr := *subnet.Ipv4CidrBlock()
			ruleNumber := 100 + i
			ruleName := "AllowOutboundToPrivate" + fmt.Sprintf("%d", i)
			publicNacl.AddEntry(jsii.String(ruleName), &awsec2.CommonNetworkAclEntryOptions{
				Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
				RuleNumber: jsii.Number(ruleNumber),
				Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1), jsii.Number(65535)),
				Direction:  awsec2.TrafficDirection_EGRESS,
			})
		}
	} else {
		// fallback based on typical CDK CIDR allocation
		privateCidr := "10.0.2.0/24"
		publicNacl.AddEntry(jsii.String("AllowOutboundToPrivate"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
			RuleNumber: jsii.Number(100),
			Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1), jsii.Number(65535)),
			Direction:  awsec2.TrafficDirection_EGRESS,
		})
	}

	// Allow outbound HTTPS to internet (for API calls, updates, etc.)
	publicNacl.AddEntry(jsii.String("AllowHTTPSOutbound"), &awsec2.CommonNetworkAclEntryOptions{
		Cidr:       awsec2.AclCidr_AnyIpv4(),
		RuleNumber: jsii.Number(110),
		Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(443)),
		Direction:  awsec2.TrafficDirection_EGRESS,
	})

	// Allow outbound HTTP for redirects and some APIs
	publicNacl.AddEntry(jsii.String("AllowHTTPOutbound"), &awsec2.CommonNetworkAclEntryOptions{
		Cidr:       awsec2.AclCidr_AnyIpv4(),
		RuleNumber: jsii.Number(120),
		Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(80)),
		Direction:  awsec2.TrafficDirection_EGRESS,
	})

	// Allow ephemeral ports outbound for return traffic
	publicNacl.AddEntry(jsii.String("AllowEphemeralOutbound"), &awsec2.CommonNetworkAclEntryOptions{
		Cidr:       awsec2.AclCidr_AnyIpv4(),
		RuleNumber: jsii.Number(130),
		Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1024), jsii.Number(65535)),
		Direction:  awsec2.TrafficDirection_EGRESS,
	})

	//* NACL for Private Subnets (AppRunner, Lambda)
	privateNacl := awsec2.NewNetworkAcl(stack, jsii.String("PrivateSubnetNacl"), &awsec2.NetworkAclProps{
		Vpc: vpc,
		SubnetSelection: &awsec2.SubnetSelection{
			Subnets: privateSubnets,
		},
		NetworkAclName: jsii.String("gyb-connect-private-nacl"),
	})

	// Private subnet NACL rules - Allow inbound from all public subnets
	if len(publicCidrs) > 0 {
		for i, publicCidr := range publicCidrs {
			ruleNumber := 100 + i
			ruleName := "AllowInboundFromPublic" + fmt.Sprintf("%d", i)
			privateNacl.AddEntry(jsii.String(ruleName), &awsec2.CommonNetworkAclEntryOptions{
				Cidr:       awsec2.AclCidr_Ipv4(&publicCidr),
				RuleNumber: jsii.Number(ruleNumber),
				Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1), jsii.Number(65535)),
				Direction:  awsec2.TrafficDirection_INGRESS,
			})
		}
	} else {
		// fallback based on typical CDK CIDR allocation
		publicCidr := "10.0.1.0/24"
		privateNacl.AddEntry(jsii.String("AllowInboundFromPublic"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&publicCidr),
			RuleNumber: jsii.Number(100),
			Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1), jsii.Number(65535)),
			Direction:  awsec2.TrafficDirection_INGRESS,
		})
	}

	// Allow outbound to all data subnets (for database access)
	if len(dataCidrs) > 0 {
		for i, dataCidr := range dataCidrs {
			ruleNumber := 100 + i
			ruleName := "AllowOutboundToDataPostgreSQL" + fmt.Sprintf("%d", i)
			privateNacl.AddEntry(jsii.String(ruleName), &awsec2.CommonNetworkAclEntryOptions{
				Cidr:       awsec2.AclCidr_Ipv4(&dataCidr),
				RuleNumber: jsii.Number(ruleNumber),
				Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(5432)), // PostgreSQL
				Direction:  awsec2.TrafficDirection_EGRESS,
			})
		}
	} else {
		// fallback based on typical CDK CIDR allocation
		dataCidr := "10.0.3.0/24"
		privateNacl.AddEntry(jsii.String("AllowOutboundToDataPostgreSQL"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&dataCidr),
			RuleNumber: jsii.Number(100),
			Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(5432)), // PostgreSQL
			Direction:  awsec2.TrafficDirection_EGRESS,
		})
	}

	// Allow outbound HTTPS to internet (for AWS API calls, external APIs) - use first private CIDR for simplicity
	if len(privateCidrs) > 0 {
		privateCidr := privateCidrs[0]
		privateNacl.AddEntry(jsii.String("AllowHTTPSOutboundPrivate"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
			RuleNumber: jsii.Number(120),
			Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(443)),
			Direction:  awsec2.TrafficDirection_EGRESS,
		})
	}

	// Allow outbound to all public subnets for responses
	if len(publicCidrs) > 0 {
		for i, publicCidr := range publicCidrs {
			ruleNumber := 130 + i
			ruleName := "AllowOutboundToPublic" + fmt.Sprintf("%d", i)
			privateNacl.AddEntry(jsii.String(ruleName), &awsec2.CommonNetworkAclEntryOptions{
				Cidr:       awsec2.AclCidr_Ipv4(&publicCidr),
				RuleNumber: jsii.Number(ruleNumber),
				Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1), jsii.Number(65535)),
				Direction:  awsec2.TrafficDirection_EGRESS,
			})
		}
	} else {
		// fallback
		publicCidr := "10.0.1.0/24"
		privateNacl.AddEntry(jsii.String("AllowOutboundToPublic"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&publicCidr),
			RuleNumber: jsii.Number(130),
			Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1), jsii.Number(65535)),
			Direction:  awsec2.TrafficDirection_EGRESS,
		})
	}

	//* Enhanced Data Subnet NACL rules (updating existing)
	dataNacl := awsec2.NewNetworkAcl(stack, jsii.String("DataSubnetNacl"), &awsec2.NetworkAclProps{
		Vpc: vpc,
		SubnetSelection: &awsec2.SubnetSelection{
			Subnets: dataSubnets,
		},
		NetworkAclName: jsii.String("gyb-connect-data-nacl"),
	})

	// Data subnet NACL rules - Allow PostgreSQL inbound from all private subnets
	if len(privateCidrs) > 0 {
		for i, privateCidr := range privateCidrs {
			ruleNumber := 100 + i
			ruleName := "AllowPostgreSQLInbound" + fmt.Sprintf("%d", i)
			dataNacl.AddEntry(jsii.String(ruleName), &awsec2.CommonNetworkAclEntryOptions{
				Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
				RuleNumber: jsii.Number(ruleNumber),
				Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(5432)), // PostgreSQL
				Direction:  awsec2.TrafficDirection_INGRESS,
			})
		}
	} else {
		// fallback
		privateCidr := "10.0.2.0/24"
		dataNacl.AddEntry(jsii.String("AllowPostgreSQLInbound"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
			RuleNumber: jsii.Number(100),
			Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(5432)), // PostgreSQL
			Direction:  awsec2.TrafficDirection_INGRESS,
		})
	}

	// Allow outbound responses to all private subnets
	if len(privateCidrs) > 0 {
		for i, privateCidr := range privateCidrs {
			ruleNumber := 100 + i
			ruleName := "AllowOutboundToPrivatePostgreSQL" + fmt.Sprintf("%d", i)
			dataNacl.AddEntry(jsii.String(ruleName), &awsec2.CommonNetworkAclEntryOptions{
				Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
				RuleNumber: jsii.Number(ruleNumber),
				Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(5432)), // PostgreSQL responses
				Direction:  awsec2.TrafficDirection_EGRESS,
			})
		}
	} else {
		// fallback
		privateCidr := "10.0.2.0/24"
		dataNacl.AddEntry(jsii.String("AllowOutboundToPrivatePostgreSQL"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
			RuleNumber: jsii.Number(100),
			Traffic:    awsec2.AclTraffic_TcpPort(jsii.Number(5432)), // PostgreSQL responses
			Direction:  awsec2.TrafficDirection_EGRESS,
		})
	}

	// Allow ephemeral ports outbound for database connection responses to all private subnets
	if len(privateCidrs) > 0 {
		for i, privateCidr := range privateCidrs {
			ruleNumber := 120 + i
			ruleName := "AllowEphemeralOutboundData" + fmt.Sprintf("%d", i)
			dataNacl.AddEntry(jsii.String(ruleName), &awsec2.CommonNetworkAclEntryOptions{
				Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
				RuleNumber: jsii.Number(ruleNumber),
				Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1024), jsii.Number(65535)),
				Direction:  awsec2.TrafficDirection_EGRESS,
			})
		}
	} else {
		// fallback
		privateCidr := "10.0.2.0/24"
		dataNacl.AddEntry(jsii.String("AllowEphemeralOutboundData"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
			RuleNumber: jsii.Number(120),
			Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1024), jsii.Number(65535)),
			Direction:  awsec2.TrafficDirection_EGRESS,
		})
	}

	// Allow ephemeral ports inbound for database connection establishment from all private subnets
	if len(privateCidrs) > 0 {
		for i, privateCidr := range privateCidrs {
			ruleNumber := 110 + i
			ruleName := "AllowEphemeralInboundData" + fmt.Sprintf("%d", i)
			dataNacl.AddEntry(jsii.String(ruleName), &awsec2.CommonNetworkAclEntryOptions{
				Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
				RuleNumber: jsii.Number(ruleNumber),
				Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1024), jsii.Number(65535)),
				Direction:  awsec2.TrafficDirection_INGRESS,
			})
		}
	} else {
		// fallback
		privateCidr := "10.0.2.0/24"
		dataNacl.AddEntry(jsii.String("AllowEphemeralInboundData"), &awsec2.CommonNetworkAclEntryOptions{
			Cidr:       awsec2.AclCidr_Ipv4(&privateCidr),
			RuleNumber: jsii.Number(110),
			Traffic:    awsec2.AclTraffic_TcpPortRange(jsii.Number(1024), jsii.Number(65535)),
			Direction:  awsec2.TrafficDirection_INGRESS,
		})
	}

	// VPC Core Outputs
	awscdk.NewCfnOutput(stack, jsii.String("VpcId"), &awscdk.CfnOutputProps{
		Value:       vpc.VpcId(),
		Description: jsii.String("ID of the VPC"),
		ExportName:  jsii.String("GybConnect-VpcId"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("VpcCidr"), &awscdk.CfnOutputProps{
		Value:       vpc.VpcCidrBlock(),
		Description: jsii.String("CIDR block of the VPC"),
		ExportName:  jsii.String("GybConnect-VpcCidr"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("VpcArn"), &awscdk.CfnOutputProps{
		Value:       vpc.VpcArn(),
		Description: jsii.String("ARN of the VPC"),
		ExportName:  jsii.String("GybConnect-VpcArn"),
	})

	// Public Subnet Outputs
	if len(*publicSubnets) > 0 {
		publicSubnetSelection := vpc.SelectSubnets(&awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PUBLIC})
		awscdk.NewCfnOutput(stack, jsii.String("PublicSubnetIds"), &awscdk.CfnOutputProps{
			Value:       awscdk.Fn_Join(jsii.String(","), publicSubnetSelection.SubnetIds),
			Description: jsii.String("Comma-separated list of public subnet IDs"),
			ExportName:  jsii.String("GybConnect-PublicSubnetIds"),
		})

		// Join all public CIDRs
		var publicCidrPtrs []*string
		for _, cidr := range publicCidrs {
			publicCidrPtrs = append(publicCidrPtrs, jsii.String(cidr))
		}
		awscdk.NewCfnOutput(stack, jsii.String("PublicSubnetCidrs"), &awscdk.CfnOutputProps{
			Value:       awscdk.Fn_Join(jsii.String(","), &publicCidrPtrs),
			Description: jsii.String("CIDR blocks of all public subnets"),
			ExportName:  jsii.String("GybConnect-PublicSubnetCidrs"),
		})
	}

	// Private Subnet Outputs
	if len(*privateSubnets) > 0 {
		privateSubnetSelection := vpc.SelectSubnets(&awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS})
		awscdk.NewCfnOutput(stack, jsii.String("PrivateSubnetIds"), &awscdk.CfnOutputProps{
			Value:       awscdk.Fn_Join(jsii.String(","), privateSubnetSelection.SubnetIds),
			Description: jsii.String("Comma-separated list of private subnet IDs"),
			ExportName:  jsii.String("GybConnect-PrivateSubnetIds"),
		})

		// Join all private CIDRs
		var privateCidrPtrs []*string
		for _, cidr := range privateCidrs {
			privateCidrPtrs = append(privateCidrPtrs, jsii.String(cidr))
		}
		awscdk.NewCfnOutput(stack, jsii.String("PrivateSubnetCidrs"), &awscdk.CfnOutputProps{
			Value:       awscdk.Fn_Join(jsii.String(","), &privateCidrPtrs),
			Description: jsii.String("CIDR blocks of all private subnets"),
			ExportName:  jsii.String("GybConnect-PrivateSubnetCidrs"),
		})
	}

	// Data/Isolated Subnet Outputs
	if len(*dataSubnets) > 0 {
		dataSubnetSelection := vpc.SelectSubnets(&awsec2.SubnetSelection{SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED})
		awscdk.NewCfnOutput(stack, jsii.String("DataSubnetIds"), &awscdk.CfnOutputProps{
			Value:       awscdk.Fn_Join(jsii.String(","), dataSubnetSelection.SubnetIds),
			Description: jsii.String("Comma-separated list of data/isolated subnet IDs"),
			ExportName:  jsii.String("GybConnect-DataSubnetIds"),
		})

		// Join all data CIDRs
		var dataCidrPtrs []*string
		for _, cidr := range dataCidrs {
			dataCidrPtrs = append(dataCidrPtrs, jsii.String(cidr))
		}
		awscdk.NewCfnOutput(stack, jsii.String("DataSubnetCidrs"), &awscdk.CfnOutputProps{
			Value:       awscdk.Fn_Join(jsii.String(","), &dataCidrPtrs),
			Description: jsii.String("CIDR blocks of all data/isolated subnets"),
			ExportName:  jsii.String("GybConnect-DataSubnetCidrs"),
		})
	}

	// Availability Zones Output
	awscdk.NewCfnOutput(stack, jsii.String("AvailabilityZones"), &awscdk.CfnOutputProps{
		Value:       awscdk.Fn_Join(jsii.String(","), vpc.AvailabilityZones()),
		Description: jsii.String("Comma-separated list of availability zones"),
		ExportName:  jsii.String("GybConnect-AvailabilityZones"),
	})

	// NACL Outputs for Security Reference
	awscdk.NewCfnOutput(stack, jsii.String("PublicNaclId"), &awscdk.CfnOutputProps{
		Value:       publicNacl.NetworkAclId(),
		Description: jsii.String("Network ACL ID for public subnets"),
		ExportName:  jsii.String("GybConnect-PublicNaclId"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("PrivateNaclId"), &awscdk.CfnOutputProps{
		Value:       privateNacl.NetworkAclId(),
		Description: jsii.String("Network ACL ID for private subnets"),
		ExportName:  jsii.String("GybConnect-PrivateNaclId"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("DataNaclId"), &awscdk.CfnOutputProps{
		Value:       dataNacl.NetworkAclId(),
		Description: jsii.String("Network ACL ID for data subnets"),
		ExportName:  jsii.String("GybConnect-DataNaclId"),
	})

	// Internet Gateway Output
	awscdk.NewCfnOutput(stack, jsii.String("InternetGatewayId"), &awscdk.CfnOutputProps{
		Value:       vpc.InternetGatewayId(),
		Description: jsii.String("Internet Gateway ID"),
		ExportName:  jsii.String("GybConnect-InternetGatewayId"),
	})

	// VPC Endpoints Output (for future VPC endpoints)
	awscdk.NewCfnOutput(stack, jsii.String("VpcEndpointSecurityGroup"), &awscdk.CfnOutputProps{
		Value:       jsii.String("placeholder-for-vpc-endpoints-sg"),
		Description: jsii.String("Security Group ID for VPC Endpoints (to be implemented)"),
		ExportName:  jsii.String("GybConnect-VpcEndpointSecurityGroup"),
	})

	// PCI DSS Requirement 10.2: VPC Flow Logs Outputs
	awscdk.NewCfnOutput(stack, jsii.String("VpcFlowLogsBucketName"), &awscdk.CfnOutputProps{
		Value:       vpcFlowLogsBucket.BucketName(),
		Description: jsii.String("S3 Bucket for VPC Flow Logs long-term storage"),
		ExportName:  jsii.String("GybConnect-VpcFlowLogsBucket"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("VpcFlowLogsBucketArn"), &awscdk.CfnOutputProps{
		Value:       vpcFlowLogsBucket.BucketArn(),
		Description: jsii.String("ARN of the S3 Bucket for VPC Flow Logs"),
		ExportName:  jsii.String("GybConnect-VpcFlowLogsBucketArn"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("VpcFlowLogsGroupName"), &awscdk.CfnOutputProps{
		Value:       vpcFlowLogsGroup.LogGroupName(),
		Description: jsii.String("CloudWatch Log Group for VPC Flow Logs"),
		ExportName:  jsii.String("GybConnect-VpcFlowLogsGroup"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("VpcFlowLogsGroupArn"), &awscdk.CfnOutputProps{
		Value:       vpcFlowLogsGroup.LogGroupArn(),
		Description: jsii.String("ARN of the CloudWatch Log Group for VPC Flow Logs"),
		ExportName:  jsii.String("GybConnect-VpcFlowLogsGroupArn"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("VpcFlowLogsRoleArn"), &awscdk.CfnOutputProps{
		Value:       flowLogsRole.RoleArn(),
		Description: jsii.String("IAM Role ARN for VPC Flow Logs"),
		ExportName:  jsii.String("GybConnect-VpcFlowLogsRoleArn"),
	})

	return &VpcStack{
		Stack: stack,
		Vpc:   vpc,
	}
}
