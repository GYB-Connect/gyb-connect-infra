package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
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
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// VPC for shared infrastructure
	vpc := awsec2.NewVpc(stack, jsii.String("GybConnectVpc"), &awsec2.VpcProps{
		MaxAzs:     jsii.Number(2),
		VpcName:    jsii.String("gyb-connect-vpc"),
		EnableDnsHostnames: jsii.Bool(true),
		EnableDnsSupport:   jsii.Bool(true),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// Output VPC ID for cross-stack references
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

	return &VpcStack{
		Stack: stack,
		Vpc:   vpc,
	}
}
