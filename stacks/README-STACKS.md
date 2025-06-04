# GYB Connect Infrastructure - Modular Stack Architecture

This CDK project implements the GYB Connect infrastructure using a modular approach with separate stacks for each service. This design follows AWS CDK best practices for maintainability, security, and independent deployment.

## Architecture Overview

The infrastructure is divided into the following stacks:

### 1. VPC Stack (`stacks/vpc_stack.go`)
- **Purpose**: Foundational networking infrastructure
- **Resources**: 
  - VPC with public and private subnets across 2 AZs
  - Internet Gateway and NAT Gateways
  - Route tables and security groups base
- **Dependencies**: None
- **Stack Name**: `GybConnect-VpcStack`

### 2. S3 Stack (`stacks/s3_stack.go`)
- **Purpose**: File storage for uploads
- **Resources**:
  - S3 bucket with versioning enabled
  - Server-side encryption (S3 managed)
  - CORS configuration for web uploads
  - Lifecycle rules for cost optimization
- **Dependencies**: None
- **Stack Name**: `GybConnect-S3Stack`

### 3. DynamoDB Stack (`stacks/dynamodb_stack.go`)
- **Purpose**: User logs and application data
- **Resources**:
  - DynamoDB table with partition key (userId) and sort key (timestamp)
  - Global Secondary Index for querying by action type
  - Encryption at rest and point-in-time recovery
  - TTL attribute for automatic cleanup
- **Dependencies**: None
- **Stack Name**: `GybConnect-DynamoDBStack`

### 4. RDS Stack (`stacks/rds_stack.go`)
- **Purpose**: PostgreSQL database for relational data
- **Resources**:
  - RDS PostgreSQL 15.4 instance (t3.micro)
  - Security group with VPC-only access
  - Automated backups and performance insights
  - Secrets Manager integration for credentials
- **Dependencies**: VPC Stack
- **Stack Name**: `GybConnect-RDSStack`

### 5. API Gateway Stack (`stacks/apigateway_stack.go`)

- **Purpose**: HTTP API endpoints and management
- **Resources**:
  - HTTP API with CORS configuration
  - Auto-deploying stage with throttling
  - CloudWatch integration ready
  - Placeholder routes for uploads and logs
- **Dependencies**: None (can reference other stacks via exports)
- **Stack Name**: `GybConnect-ApiGatewayStack`

## Deployment Strategy

### Development Environment
```bash
# Deploy all stacks
cdk deploy --all

# Deploy specific stack
cdk deploy GybConnect-S3Stack

# Deploy in dependency order
cdk deploy GybConnect-VpcStack
cdk deploy GybConnect-RDSStack
```

### Production Considerations

Before deploying to production, update the following settings:

1. **S3 Stack**:
   - Set `RemovalPolicy` to `RETAIN`
   - Set `AutoDeleteObjects` to `false`
   - Restrict CORS origins to your actual domains

2. **RDS Stack**:
   - Set `DeletionProtection` to `true`
   - Set `RemovalPolicy` to `RETAIN`
   - Consider larger instance types
   - Review backup retention settings

3. **DynamoDB Stack**:
   - Set `RemovalPolicy` to `RETAIN`
   - Consider provisioned billing for predictable workloads

4. **API Gateway Stack**:
   - Update CORS origins to your actual frontend URLs
   - Implement proper authentication/authorization
   - Review throttling settings for your expected load

## Cross-Stack References

The stacks use CloudFormation exports to share resources:

- `GybConnect-VpcId`: VPC ID for cross-stack references
- `GybConnect-S3BucketName`: S3 bucket name
- `GybConnect-S3BucketArn`: S3 bucket ARN
- `GybConnect-DynamoDBTableName`: DynamoDB table name
- `GybConnect-DynamoDBTableArn`: DynamoDB table ARN
- `GybConnect-DatabaseEndpoint`: RDS endpoint
- `GybConnect-DatabaseSecretArn`: Database credentials secret ARN
- `GybConnect-HttpApiUrl`: HTTP API Gateway URL
- `GybConnect-HttpApiId`: HTTP API Gateway ID

## Next Steps

1. **Lambda Functions**: Create separate Lambda stacks for business logic
2. **Authentication**: Add Cognito User Pool for user management
3. **Monitoring**: Add CloudWatch dashboards and alarms
4. **CI/CD**: Set up deployment pipelines
5. **Security**: Implement WAF and additional security measures

## File Structure

```
gyb-connect-infra/
├── gyb_connect.go          # Main orchestration file
├── stacks/
│   ├── vpc_stack.go        # VPC and networking
│   ├── s3_stack.go         # S3 file storage
│   ├── dynamodb_stack.go   # DynamoDB user logs
│   ├── rds_stack.go        # PostgreSQL database
│   └── apigateway_stack.go # API Gateway
├── go.mod
├── go.sum
├── cdk.json
└── README.md
```

## Benefits of This Architecture

1. **Independent Deployment**: Each service can be deployed separately
2. **Better Testing**: Test individual components in isolation
3. **Security Boundaries**: Separate IAM roles and policies per stack
4. **Cost Management**: Track costs per service
5. **Team Collaboration**: Different teams can own different stacks
6. **Environment Management**: Deploy different combinations for dev/staging/prod
7. **Rollback Safety**: Rollback individual services without affecting others
