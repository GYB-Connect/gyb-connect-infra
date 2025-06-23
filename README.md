# GYB Connect Infrastructure - PCI DSS Compliant

This CDK project implements the AWS infrastructure for GYB Connect using a **modular stack architecture** following AWS CDK best practices and **PCI DSS SAQ D-SP compliance requirements**.

## 🏗️ Architecture Overview

The infrastructure is split into independent, deployable stacks:

- **KMS Stack**: Customer-Managed Encryption Keys for PCI DSS compliance
- **Security Stack**: GuardDuty, Inspector, and Security Hub for threat detection
- **VPC Stack**: Networking foundation (VPC, subnets, gateways)
- **S3 Stack**: File storage with customer-managed encryption
- **DynamoDB Stack**: User logs database with customer-managed encryption
- **IAM Stack**: Least privilege roles and MFA enforcement policies
- **Logging Stack**: Centralized logging with CloudTrail and real-time alerting
- **RDS Stack**: PostgreSQL database with SSL enforcement and encryption
- **API Gateway Stack**: HTTPS API endpoints with TLS 1.2+ enforcement

For detailed information about each stack, see [README-STACKS.md](./README-STACKS.md).

## 🔐 PCI DSS Compliance Features

This infrastructure implements controls for PCI DSS Requirements:

- **Req 2**: Secure configurations with environment-specific settings
- **Req 3**: Customer-managed encryption keys (CMK) for data at rest
- **Req 4**: TLS 1.2+ enforcement and SSL for database connections
- **Req 5**: GuardDuty malware protection
- **Req 6**: Security vulnerability scanning with Inspector
- **Req 7**: Least privilege access control with role-based permissions
- **Req 8**: Strong authentication with mandatory MFA enforcement
- **Req 10**: Comprehensive logging and monitoring with CloudTrail and real-time alerts
- **Req 11**: Continuous security monitoring with Security Hub

See the `docs/` directory for detailed compliance documentation.

## 🚀 Quick Start

### Prerequisites

- AWS CLI configured with appropriate credentials
- AWS CDK v2 installed: `npm install -g aws-cdk`
- Go 1.19+ installed
- ACM Certificate for custom domains (for API Gateway TLS)

### Environment Setup

```bash
# Set up your ACM certificate ARN
export ACM_CERTIFICATE_ARN="arn:aws:acm:us-west-1:123345667789:certificate/1234567890"

# Choose your environment
export DEPLOY_ENV="dev"  # or "prod" for production
```

### Deployment

The infrastructure automatically deploys stacks in the correct order:

```bash
# Deploy all stacks for development
./deploy.sh dev

# Deploy all stacks for production (with confirmation)
./deploy.sh prod

# Individual stack deployments (rarely needed)
cdk deploy GybConnect-KmsStack
cdk deploy GybConnect-SecurityStack
```

#### Deployment Order

The stacks deploy in this specific order to handle dependencies:

```
1. KMS Stack (encryption keys)
2. Security Stack (monitoring services)
3. VPC Stack (networking)
4. S3 Stack (uses KMS key)
5. DynamoDB Stack (uses KMS key)
6. IAM Stack (uses S3 and DynamoDB resources)
7. Logging Stack (uses KMS key)
8. RDS Stack (uses VPC and KMS key)
9. API Gateway Stack (uses certificate)
```

## 📋 Management Commands

Use the provided scripts for common operations:

```bash
# Deployment
./deploy.sh dev          # Deploy development environment
./deploy.sh prod         # Deploy production environment

# Management operations  
./manage.sh status       # Check what's deployed
./manage.sh outputs      # View stack outputs
./manage.sh diff         # See pending changes
./manage.sh validate     # Validate configuration
./manage.sh destroy      # Destroy all stacks (be careful!)
```

## 🔧 Development

```bash
# Validate syntax
go build .

# Synthesize CloudFormation
cdk synth --all

# View specific stack template
cdk synth GybConnect-SecurityStack

# Check differences
cdk diff --all
```

## 🌍 Environment Configuration

**Environment Differences:**

| Feature | Development | Production |
|---------|-------------|------------|
| VPC | AWS Default VPC | Custom VPC Stack |
| Deletion Protection | Disabled | Enabled |
| Removal Policy | DESTROY | RETAIN |
| CORS Origins | localhost, dev domains | Production domains only |
| Custom Domain | api-dev.gybconnect.com | api.gybconnect.com |
| Security Alerts | Optional | Required |

## 📊 Stack Dependencies

```
KMS Stack (foundational - encryption keys)
├── S3 Stack (uses S3 KMS key)
├── DynamoDB Stack (uses DynamoDB KMS key)
├── IAM Stack (uses S3/DynamoDB resources)
│   └── Requires S3 and DynamoDB stacks
├── Logging Stack (uses Logging KMS key)
│   └── Requires KMS stack
├── RDS Stack (uses RDS KMS key)
│   └── Requires VPC Stack
└── Security Stack (independent monitoring)

API Gateway Stack (uses ACM certificate)
```

## 🔐 Security Features

### Encryption

- **KMS**: Customer-managed keys with automatic rotation
- **S3**: CMK encryption for all objects
- **DynamoDB**: CMK encryption for tables
- **RDS**: CMK encryption + SSL enforcement

### Monitoring & Detection

- **GuardDuty**: Malware and threat detection
- **Inspector**: Vulnerability scanning
- **Security Hub**: Compliance monitoring
- **EventBridge**: Automated security alerts

### Network Security

- **VPC**: Private subnets for databases
- **Security Groups**: Least privilege access
- **API Gateway**: TLS 1.2+ only
- **WAF**: Protection against common attacks

### Access Control

- **IAM Identity Center**: Centralized authentication with MFA
- **Least Privilege Roles**: Specific roles for each service
- **MFA Enforcement**: Boundary policies require MFA for all actions
- **IAM Access Analyzer**: Continuous permission monitoring

### Logging and Monitoring

- **CloudTrail**: Multi-region API logging with file validation
- **CloudWatch Logs**: Real-time log processing and retention
- **S3 Central Logging**: Immutable storage with Object Lock
- **EventBridge Rules**: Real-time security event monitoring
- **SNS Alerts**: Immediate notification of critical events



## 🚨 Production Checklist

Before deploying to production:

1. ✅ **Set certificate ARN** in environment variable
2. ✅ **Update domain names** in `gyb_connect.go`
3. ✅ **Configure security alerts** email in `security_stack.go`
4. ✅ **Review instance types** for production workloads
5. ✅ **Set up DNS records** for custom domains
6. ✅ **Enable IAM Identity Center** with MFA for all users
7. ✅ **Configure IAM Access Analyzer** for continuous monitoring
8. ✅ **Configure security alert email** in logging stack
9. ✅ **Verify CloudTrail is logging** and alerts are working
10. ✅ **Enable Amazon Macie** for data discovery
11. ✅ **Schedule ASV scans** for PCI compliance
12. ✅ **Plan penetration testing** annually

## 📚 Documentation

### Compliance Guides

- [PCI DSS Compliance Roadmap](docs/PCI%20DSS%20SAQ%20D-SP%20Compliance%20Roadmap%20for%20GYB%20Connect.md)
- [Certificate Integration Guide](docs/Certificate_Integration_Guide.md)
- [Requirements 5, 6, 11 Summary](docs/PCI_DSS_Requirements_5_6_11_Summary.md)
- [Requirements 7 & 8 Summary](docs/PCI_DSS_Requirements_7_8_Summary.md)
- [Requirement 10 Summary](docs/PCI_DSS_Requirement_10_Summary.md)

### Implementation Guides

- [Requirement 6: Secure SDLC](docs/PCI_DSS_Requirement_6_SDLC_Security_Guide.md)
- [Requirements 7 & 8: Access Control & Authentication](docs/PCI_DSS_Requirements_7_8_Implementation_Guide.md)
- [Requirement 10: Logging & Monitoring](docs/PCI_DSS_Requirement_10_Implementation_Guide.md)
- [Requirement 11: Security Testing](docs/PCI_DSS_Requirement_11_Testing_Guide.md)

### Technical Documentation

- [Detailed Stack Documentation](./README-STACKS.md)
- [AWS CDK Developer Guide](https://docs.aws.amazon.com/cdk/latest/guide/)
- [AWS Well-Architected Framework](https://aws.amazon.com/architecture/well-architected/)

## 🤝 Contributing

1. Follow the modular stack pattern
2. Add PCI DSS requirement comments in code
3. Update documentation for new features
4. Test with `./manage.sh validate` before committing
5. Use meaningful stack and resource names
6. Ensure security best practices

## 📞 Support

For security issues or compliance questions:

- Security Team: <security@gybconnect.com>
- Infrastructure: <devops@gybconnect.com>
