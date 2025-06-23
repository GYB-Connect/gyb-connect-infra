package stacks

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigatewayv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awscertificatemanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type ApiGatewayStackProps struct {
	awscdk.StackProps
	Environment string
	// PCI DSS Req 4.1: Accept custom domain name for TLS enforcement
	DomainName string
	// PCI DSS Req 4.1: Accept ACM certificate for custom domain
	Certificate awscertificatemanager.ICertificate
}

type ApiGatewayStack struct {
	awscdk.Stack
	HttpApi      awsapigatewayv2.HttpApi
	CustomDomain awsapigatewayv2.IDomainName
}

// NewApiGatewayStack creates an API Gateway with PCI DSS compliant security settings
// This stack implements controls for PCI DSS Requirements 1.2, 2.2, 4.1, and 6.4.2
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

	// PCI DSS Req 2.2 & 6.4.2: Determine CORS origins based on environment
	// Production environments must restrict access to specific, known domains only
	var corsOrigins []*string
	if envPrefix == PROD_ENV {
		// Production origins - strictly limited to production domains
		// PCI DSS Req 1.2: Restrict inbound traffic to only necessary sources
		corsOrigins = []*string{
			jsii.String("https://app.gybconnect.com"), // Replace with your actual production domain
			jsii.String("https://www.gybconnect.com"),
		}
	} else {
		// Development origins - broader access for testing and development
		// Still avoiding wildcards to establish good security practices
		corsOrigins = []*string{
			jsii.String("http://localhost:3000"),
			jsii.String("http://localhost:5173"), // Vite default
			jsii.String("https://localhost:3000"), // HTTPS local development
			jsii.String("https://dev.gybconnect.com"), // Replace with your actual dev domain
		}
	}

	// API Gateway HTTP API
	api := awsapigatewayv2.NewHttpApi(stack, jsii.String("GybConnectHttpApi"), &awsapigatewayv2.HttpApiProps{
		ApiName:     jsii.String(apiName),
		Description: jsii.String("HTTP API Gateway for GYB Connect application"),

		// PCI DSS Req 2.2 & 6.4.2: CORS configuration with security best practices
		// Implements defense-in-depth by restricting cross-origin requests
		CorsPreflight: &awsapigatewayv2.CorsPreflightOptions{
			// PCI DSS Req 6.4.2: Only allow specific, trusted origins
			// Never use wildcard (*) origins in production environments
			AllowOrigins: &corsOrigins,
			// PCI DSS Req 2.2: Define only the HTTP methods required by the application
			// Following principle of least privilege
			AllowMethods: &[]awsapigatewayv2.CorsHttpMethod{
				awsapigatewayv2.CorsHttpMethod_GET,
				awsapigatewayv2.CorsHttpMethod_POST,
				awsapigatewayv2.CorsHttpMethod_PUT,
				awsapigatewayv2.CorsHttpMethod_DELETE,
				awsapigatewayv2.CorsHttpMethod_OPTIONS,
			},
			// TODO: PCI DSS Req 2.2: Consider restricting to specific headers instead of wildcard
			// For production, enumerate only the headers your application actually uses
			AllowHeaders: &[]*string{
				jsii.String("*"),
			},
			// PCI DSS Req 8.1: Enable credentials for secure authentication flows
			AllowCredentials: jsii.Bool(true),
		},
	})

	// Create a default stage
	// PCI DSS Req 10.2: Stages help with logging and monitoring of API access
	stage := awsapigatewayv2.NewHttpStage(stack, jsii.String("GybConnectHttpApiStage"), &awsapigatewayv2.HttpStageProps{
		HttpApi:    api,
		StageName:  jsii.String(envPrefix), // Use environment-specific stage names
		// PCI DSS Req 6.5: Auto-deploy enables immediate security updates
		AutoDeploy: jsii.Bool(true),
	})

	// PCI DSS Req 4.1: Configure custom domain with TLS 1.2+ if domain and certificate are provided
	var customDomain awsapigatewayv2.IDomainName
	if props != nil && props.DomainName != "" && props.Certificate != nil {
		customDomain = awsapigatewayv2.NewDomainName(stack, jsii.String("GybConnectCustomDomain"), &awsapigatewayv2.DomainNameProps{
			DomainName:  jsii.String(props.DomainName),
			Certificate: props.Certificate,
			// PCI DSS Req 4.1: Enforce TLS 1.2+ - reject connections using older protocols
			SecurityPolicy: awsapigatewayv2.SecurityPolicy_TLS_1_2,
		})

		// Associate the custom domain with the API stage
		awsapigatewayv2.NewApiMapping(stack, jsii.String("GybConnectApiMapping"), &awsapigatewayv2.ApiMappingProps{
			Api:        api,
			DomainName: customDomain,
			Stage:      stage,
		})

		// Output custom domain information
		awscdk.NewCfnOutput(stack, jsii.String("CustomDomainName"), &awscdk.CfnOutputProps{
			Value:       customDomain.Name(),
			Description: jsii.String("Custom domain name for the API"),
			ExportName:  jsii.String("GybConnect-CustomDomainName-" + envPrefix),
		})

		awscdk.NewCfnOutput(stack, jsii.String("CustomDomainAlias"), &awscdk.CfnOutputProps{
			Value:       customDomain.RegionalDomainName(),
			Description: jsii.String("Regional domain name alias for DNS configuration"),
			ExportName:  jsii.String("GybConnect-CustomDomainAlias-" + envPrefix),
		})
	}

	// Note: HTTP API routes will be added when Lambda functions are integrated
	// PCI DSS Req 7.1: Routes should implement least privilege access control
	// PCI DSS Req 8.1: Each route should have appropriate authentication/authorization
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
		ExportName:  jsii.String("GybConnect-HttpApiUrl-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("HttpApiId"), &awscdk.CfnOutputProps{
		Value:       api.ApiId(),
		Description: jsii.String("ID of the HTTP API Gateway"),
		ExportName:  jsii.String("GybConnect-HttpApiId-" + envPrefix),
	})

	awscdk.NewCfnOutput(stack, jsii.String("HttpApiStage"), &awscdk.CfnOutputProps{
		Value:       stage.StageName(),
		Description: jsii.String("Name of the HTTP API Stage"),
		ExportName:  jsii.String("GybConnect-HttpApiStage-" + envPrefix),
	})

	return &ApiGatewayStack{
		Stack:        stack,
		HttpApi:      api,
		CustomDomain: customDomain,
	}
}
