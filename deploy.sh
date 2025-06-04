#!/bin/bash

# GYB Connect Infrastructure Deployment Script
# This script deploys the CDK stacks in the correct dependency order

set -e  # Exit on any error

# Set AWS profile and region
export AWS_PROFILE=gyb-connect
export AWS_DEFAULT_REGION=us-west-1

echo "🚀 Starting GYB Connect Infrastructure Deployment"
echo "=================================================="
echo "Using AWS Profile: $AWS_PROFILE"
echo "Using AWS Region: $AWS_DEFAULT_REGION"

# Check if CDK is installed
if ! command -v cdk &> /dev/null; then
    echo "❌ AWS CDK is not installed. Please install it first:"
    echo "npm install -g aws-cdk"
    exit 1
fi

# Check if AWS credentials are configured for the profile
if ! aws sts get-caller-identity --profile $AWS_PROFILE &> /dev/null; then
    echo "❌ AWS credentials are not configured for profile '$AWS_PROFILE'."
    echo "Please run: aws configure --profile $AWS_PROFILE"
    exit 1
fi

# Display current AWS account and region
echo "📋 AWS Account Information:"
aws sts get-caller-identity --profile $AWS_PROFILE --output table

# Function to deploy a stack with retry logic
deploy_stack() {
    local stack_name=$1
    local description=$2
    
    echo ""
    echo "📦 Deploying $description..."
    echo "Stack: $stack_name"
    echo "----------------------------------------"
    
    if cdk deploy $stack_name --require-approval never --profile $AWS_PROFILE; then
        echo "✅ Successfully deployed $stack_name"
    else
        echo "❌ Failed to deploy $stack_name"
        exit 1
    fi
}

# Parse command line arguments
DEPLOY_ALL=false
STACK_TO_DEPLOY=""
ENVIRONMENT="dev"  #* Default to development

case "${1:-all}" in
    "all")
        DEPLOY_ALL=true
        ;;
    "all-prod")
        DEPLOY_ALL=true
        ENVIRONMENT="production"
        ;;
    "vpc")
        STACK_TO_DEPLOY="GybConnect-VpcStack"
        ;;
    "s3")
        STACK_TO_DEPLOY="GybConnect-S3Stack"
        ;;
    "dynamodb")
        STACK_TO_DEPLOY="GybConnect-DynamoDBStack"
        ;;
    "rds")
        STACK_TO_DEPLOY="GybConnect-RDSStack"
        ;;
    "api")
        STACK_TO_DEPLOY="GybConnect-ApiGatewayStack"
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [stack_name]"
        echo ""
        echo "Available options:"
        echo "  all      - Deploy all stacks for development (default, no VPC)"
        echo "  all-prod - Deploy all stacks for production (includes VPC)"
        echo "  vpc      - Deploy VPC stack only"
        echo "  s3       - Deploy S3 stack only"
        echo "  dynamodb - Deploy DynamoDB stack only"
        echo "  rds      - Deploy RDS stack only"
        echo "  api      - Deploy API Gateway stack only"
        echo "  help     - Show this help message"
        echo ""
        echo "Environment Variables:"
        echo "  DEPLOY_ENV - Set to 'production' for production deployment"
        echo ""
        echo "Examples:"
        echo "  $0              # Deploy all stacks for development"
        echo "  $0 all          # Deploy all stacks for development"
        echo "  $0 all-prod     # Deploy all stacks for production"
        echo "  $0 vpc          # Deploy only VPC stack"
        echo "  $0 rds          # Deploy only RDS stack"
        echo "  DEPLOY_ENV=production $0 all  # Deploy all for production using env var"
        exit 0
        ;;
    *)
        echo "❌ Unknown stack: $1"
        echo "Run '$0 help' for available options"
        exit 1
        ;;
esac

# Override environment if DEPLOY_ENV is set
if [ -n "$DEPLOY_ENV" ]; then
    ENVIRONMENT="$DEPLOY_ENV"
fi

# Export environment for CDK
export DEPLOY_ENV="$ENVIRONMENT"

echo "🌍 Deployment Environment: $ENVIRONMENT"

# Bootstrap CDK if needed
echo "🔧 Checking CDK bootstrap status..."
if ! cdk bootstrap --profile $AWS_PROFILE 2>/dev/null; then
    echo "⚠️  CDK bootstrap may be needed. This is normal for first-time deployments."
fi

if [ "$DEPLOY_ALL" = true ]; then
    echo "📋 Deploying all stacks in dependency order..."
    
    if [ "$ENVIRONMENT" = "production" ]; then
        echo "🏭 Production deployment - including VPC stack"
        # Deploy foundational stacks first (including VPC for production)
        deploy_stack "GybConnect-VpcStack" "VPC and Networking Infrastructure"
        deploy_stack "GybConnect-S3Stack" "S3 File Storage"
        deploy_stack "GybConnect-DynamoDBStack" "DynamoDB User Logs"
        
        # Deploy stacks with dependencies
        deploy_stack "GybConnect-RDSStack" "RDS PostgreSQL Database (depends on VPC)"
        deploy_stack "GybConnect-ApiGatewayStack" "API Gateway"
    else
        echo "🔧 Development deployment - using default VPC"
        # Deploy independent stacks first (no VPC needed for development)
        deploy_stack "GybConnect-S3Stack" "S3 File Storage"
        deploy_stack "GybConnect-DynamoDBStack" "DynamoDB User Logs"
        
        # Deploy RDS using default VPC
        deploy_stack "GybConnect-RDSStack" "RDS PostgreSQL Database (using default VPC)"
        deploy_stack "GybConnect-ApiGatewayStack" "API Gateway"
    fi
    
    echo ""
    echo "🎉 All stacks deployed successfully!"
    echo "=================================================="
    
    # Display important outputs
    echo ""
    echo "📋 Deployment Summary:"
    echo "======================"
    cdk list --profile $AWS_PROFILE
    
else
    deploy_stack "$STACK_TO_DEPLOY" "Individual Stack Deployment"
    echo ""
    echo "✅ Stack $STACK_TO_DEPLOY deployed successfully!"
fi

echo ""
echo "🔗 To view stack outputs, run:"
echo "   cdk synth [stack-name] --profile $AWS_PROFILE"
echo ""
echo "🗑️  To destroy all stacks, run:"
echo "   cdk destroy --all --profile $AWS_PROFILE"
echo ""
echo "📖 For more information, check README-STACKS.md"
