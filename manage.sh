#!/bin/bash

# GYB Connect Infrastructure Management Script
# Provides common operations for managing the CDK stacks

set -e

# Set AWS profile and region
export AWS_PROFILE=gyb-connect
export AWS_DEFAULT_REGION=us-west-1

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_header() {
    echo -e "${BLUE}$1${NC}"
    echo "$(printf '=%.0s' $(seq 1 ${#1}))"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Main functions
show_status() {
    print_header "Stack Status (AWS Profile: $AWS_PROFILE, Region: $AWS_DEFAULT_REGION)"
    echo "Listing all deployed stacks..."
    cdk list --deployed --profile $AWS_PROFILE 2>/dev/null || echo "No stacks deployed"
    
    echo ""
    echo "All defined stacks:"
    cdk list --profile $AWS_PROFILE
}

show_outputs() {
    print_header "Stack Outputs (AWS Profile: $AWS_PROFILE, Region: $AWS_DEFAULT_REGION)"
    
    local stacks=("GybConnect-VpcStack" "GybConnect-S3Stack" "GybConnect-DynamoDBStack" "GybConnect-RDSStack" "GybConnect-ApiGatewayStack")
    
    for stack in "${stacks[@]}"; do
        echo ""
        echo -e "${BLUE}$stack:${NC}"
        if cdk synth $stack --profile $AWS_PROFILE 2>/dev/null | grep -A 50 "Outputs:" | head -20; then
            :
        else
            echo "  No outputs or stack not found"
        fi
    done
}

show_diff() {
    local stack=${1:-"--all"}
    print_header "Stack Diff (AWS Profile: $AWS_PROFILE, Region: $AWS_DEFAULT_REGION)"
    echo "Showing differences for: $stack"
    cdk diff $stack --profile $AWS_PROFILE
}

destroy_stacks() {
    print_warning "This will destroy ALL infrastructure!"
    echo "Using AWS Profile: $AWS_PROFILE"
    echo "Using AWS Region: $AWS_DEFAULT_REGION"
    echo "The following stacks will be destroyed:"
    cdk list --profile $AWS_PROFILE
    echo ""
    read -p "Are you sure you want to continue? (type 'yes' to confirm): " confirmation
    
    if [ "$confirmation" = "yes" ]; then
        print_header "Destroying Stacks"
        # Destroy in reverse order to handle dependencies
        cdk destroy GybConnect-ApiGatewayStack --force --profile $AWS_PROFILE
        cdk destroy GybConnect-RDSStack --force --profile $AWS_PROFILE
        cdk destroy GybConnect-DynamoDBStack --force --profile $AWS_PROFILE
        cdk destroy GybConnect-S3Stack --force --profile $AWS_PROFILE
        cdk destroy GybConnect-VpcStack --force --profile $AWS_PROFILE
        print_success "All stacks destroyed"
    else
        echo "Destruction cancelled"
    fi
}

validate_stacks() {
    print_header "Stack Validation"
    echo "Validating CDK syntax and configuration..."
    
    if cdk synth --all > /dev/null; then
        print_success "All stacks validate successfully"
    else
        print_error "Stack validation failed"
        exit 1
    fi
}

bootstrap_account() {
    print_header "CDK Bootstrap"
    echo "Bootstrapping CDK in current AWS account/region..."
    cdk bootstrap
    print_success "CDK bootstrap completed"
}

show_help() {
    echo "GYB Connect Infrastructure Management"
    echo "Usage: $0 <command>"
    echo ""
    echo "Available commands:"
    echo "  status     - Show deployment status of all stacks"
    echo "  outputs    - Show stack outputs and important values"
    echo "  diff       - Show differences between deployed and local stacks"
    echo "  validate   - Validate CDK syntax and configuration"
    echo "  bootstrap  - Bootstrap CDK in current AWS account/region"
    echo "  destroy    - Destroy all stacks (with confirmation)"
    echo "  help       - Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 status          # Check what's deployed"
    echo "  $0 outputs         # See stack outputs"
    echo "  $0 diff            # See what would change"
    echo "  $0 validate        # Check if configuration is valid"
}

# Main command processing
case "${1:-help}" in
    "status")
        show_status
        ;;
    "outputs")
        show_outputs
        ;;
    "diff")
        show_diff "$2"
        ;;
    "validate")
        validate_stacks
        ;;
    "bootstrap")
        bootstrap_account
        ;;
    "destroy")
        destroy_stacks
        ;;
    "help"|"-h"|"--help")
        show_help
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_help
        exit 1
        ;;
esac
