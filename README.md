# GYB Connect Infrastructure

This CDK project implements the AWS infrastructure for GYB Connect using a **modular stack architecture** following AWS CDK best practices.

## 🏗️ Architecture Overview

The infrastructure is split into independent, deployable stacks:

- **VPC Stack**: Networking foundation (VPC, subnets, gateways)
- **S3 Stack**: File storage for uploads (`gyb-uploads-s3`)
- **DynamoDB Stack**: User logs database (`gyb-user-logs`)
- **RDS Stack**: PostgreSQL database for relational data
- **API Gateway Stack**: HTTP API endpoints with authentication

For detailed information about each stack, see [README-STACKS.md](./README-STACKS.md).

## 🚀 Quick Start

### Prerequisites

- AWS CLI configured with appropriate credentials
- AWS CDK v2 installed: `npm install -g aws-cdk`
- Go 1.23+ installed

### Deployment

The `deploy.sh` script provides environment-aware deployments with automatic dependency management:

#### Environment Options

**Development Environment (Default)**

- Uses AWS default VPC for RDS (no custom VPC needed)
- Simplified stack deployment order
- Cost-effective for development and testing

**Production Environment**

- Creates custom VPC with proper network isolation
- Enhanced security and network segmentation
- Production-ready configurations

#### Usage Examples

```bash
# DEVELOPMENT DEPLOYMENTS
./deploy.sh              # Deploy all development stacks (no VPC)
./deploy.sh all          # Same as above - explicit "all" for development

# PRODUCTION DEPLOYMENTS  
./deploy.sh all-prod     # Deploy all production stacks (includes VPC)
DEPLOY_ENV=production ./deploy.sh all  # Alternative using environment variable

# INDIVIDUAL STACK DEPLOYMENTS
./deploy.sh vpc          # VPC stack only (production environment)
./deploy.sh s3           # S3 stack only
./deploy.sh dynamodb     # DynamoDB stack only  
./deploy.sh rds          # RDS stack only (environment-aware)
./deploy.sh api          # API Gateway stack only

# ENVIRONMENT OVERRIDE
DEPLOY_ENV=production ./deploy.sh rds   # Deploy RDS with production VPC
DEPLOY_ENV=development ./deploy.sh rds  # Deploy RDS with default VPC
```

#### Deployment Flow

**Development Flow (./deploy.sh all)**

```
1. S3 Stack
2. DynamoDB Stack  
3. RDS Stack (default VPC)
4. API Gateway Stack
```

**Production Flow (./deploy.sh all-prod)**

```
1. VPC Stack (foundation)
2. S3 Stack
3. DynamoDB Stack
4. RDS Stack (custom VPC)
5. API Gateway Stack
```

## 📋 Management Commands

Use the provided scripts for common operations:

```bash
# Environment-aware deployment (see above for detailed examples)
./deploy.sh [all|all-prod|vpc|s3|dynamodb|rds|api]

# Management operations  
./manage.sh status       # Check what's deployed
./manage.sh outputs      # View stack outputs
./manage.sh diff         # See pending changes
./manage.sh validate     # Validate configuration
./manage.sh destroy      # Destroy all stacks
```

## 🔧 Development

```bash
# Validate syntax
go build .

# Synthesize CloudFormation
cdk synth --all

# View specific stack template
cdk synth GybConnect-S3Stack

# Check differences
cdk diff --all
```

## 🌍 Environment Detection

The infrastructure automatically detects the target environment using the following priority order:

1. **CDK Context**: `cdk deploy -c environment=production`
2. **Environment Variable**: `DEPLOY_ENV=production`  
3. **Default**: Development environment (uses default VPC)

**Environment Differences:**

| Feature | Development | Production |
|---------|-------------|------------|
| VPC | AWS Default VPC | Custom VPC Stack |
| RDS Networking | Default VPC subnets | Private subnets |
| Security Groups | Default VPC security group | Custom security groups |
| Deployment Order | S3 → DynamoDB → RDS → API | VPC → S3 → DynamoDB → RDS → API |
| Cost | Lower (no VPC costs) | Higher (VPC NAT, etc.) |

The stacks are environment-agnostic by default. To deploy to specific environments:

1. Update the `env()` function in `gyb_connect.go`
2. Set environment-specific configurations in each stack

## 📊 Stack Dependencies

```
VPC Stack (foundational)
├── RDS Stack (requires VPC)
└── Independent stacks:
    ├── S3 Stack
    ├── DynamoDB Stack
    └── API Gateway Stack
```

## 🔐 Security Features

- **S3**: Bucket encryption, versioning, block public access
- **DynamoDB**: Encryption at rest, TTL for automatic cleanup  
- **RDS**: VPC isolation, security groups, automated backups
- **API Gateway**: CORS configuration, API keys, throttling

## 💰 Cost Optimization

- S3 lifecycle rules for incomplete uploads
- DynamoDB pay-per-request billing
- RDS t3.micro instance (free tier eligible)
- Automatic cleanup configurations

## 🚨 Production Checklist

Before deploying to production:

1. **Update removal policies** to `RETAIN` for data stores
2. **Configure proper CORS origins** for API Gateway
3. **Enable deletion protection** for RDS
4. **Review instance types** and scaling settings
5. **Set up monitoring** and alerting
6. **Configure backup strategies**

## 📚 Additional Resources

- [Detailed Stack Documentation](./README-STACKS.md)
- [AWS CDK Developer Guide](https://docs.aws.amazon.com/cdk/latest/guide/)
- [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/)

## 🤝 Contributing

1. Follow the modular stack pattern
2. Update documentation for new stacks
3. Test with `./manage.sh validate` before committing
4. Use meaningful stack and resource names
