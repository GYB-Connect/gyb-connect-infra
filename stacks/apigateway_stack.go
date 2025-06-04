package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ApiGatewayStackProps struct {
	awscdk.StackProps
	Environment string
}

type ApiGatewayStack struct {
	awscdk.Stack
	HttpApi awsapigatewayv2.HttpApi
}

func NewApiGatewayStack(scope constructs.Construct, id string, props *ApiGatewayStackProps) *ApiGatewayStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	envPrefix := DEV_ENV // Default to development environment
	if props != nil && props.Environment != "" {
		envPrefix = props.Environment
	}
	apiName := envPrefix + "-gyb-connect-api"

	// API Gateway HTTP API
	api := awsapigatewayv2.NewHttpApi(stack, jsii.String("GybConnectHttpApi"), &awsapigatewayv2.HttpApiProps{
		ApiName:     jsii.String(apiName),
		Description: jsii.String("HTTP API Gateway for GYB Connect application"),

		// CORS configuration
		CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
			AllowOrigins: &[]*string{
				jsii.String("https://localhost:3000"), // Add your frontend URLs
				jsii.String("https://yourdomain.com"),
			},
			AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{
				awsapigatewayv2.CorsHttpMethod_GET,
				awsapigatewayv2.CorsHttpMethod_POST,
				awsapigatewayv2.CorsHttpMethod_PUT,
				awsapigatewayv2.CorsHttpMethod_DELETE,
				awsapigatewayv2.CorsHttpMethod_OPTIONS,
			},
			AllowHeaders: &[]*string{
				jsii.String("*"),
			},
			AllowCredentials: jsii.Bool(true),
		},
	})

	// Create a default stage
	stage := awsapigatewayv2.NewHttpStage(stack, jsii.String("GybConnectHttpApiStage"), &awsapigatewayv2.HttpStageProps{
		HttpApi:    api,
		StageName:  jsii.String("prod"),
		AutoDeploy: jsii.Bool(true),
	})

	// Note: HTTP API routes will be added when Lambda functions are integrated
	// For now, we'll skip creating routes since they require integrations
	// Routes can be added later using:
	// api.AddRoutes(&awsapigatewayv2.AddRoutesOptions{
	//     Path: jsii.String("/uploads"),
	//     Methods: &[]awsapigatewayv2.HttpMethod{awsapigatewayv2.HttpMethod_GET, awsapigatewayv2.HttpMethod_POST},
	//     Integration: someIntegration,
	// })

	// Output API information
	awscdk.NewCfnOutput(stack, jsii.String("HttpApiUrl"), &awscdk.CfnOutputProps{
		Value:       api.ApiEndpoint(),
		Description: jsii.String("URL of the HTTP API Gateway"),
		ExportName:  jsii.String("GybConnect-HttpApiUrl"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("HttpApiId"), &awscdk.CfnOutputProps{
		Value:       api.ApiId(),
		Description: jsii.String("ID of the HTTP API Gateway"),
		ExportName:  jsii.String("GybConnect-HttpApiId"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("HttpApiStage"), &awscdk.CfnOutputProps{
		Value:       stage.StageName(),
		Description: jsii.String("Name of the HTTP API Stage"),
		ExportName:  jsii.String("GybConnect-HttpApiStage"),
	})

	return &ApiGatewayStack{
		Stack:   stack,
		HttpApi: api,
	}
}
