# PCI DSS Requirements 7 & 8 Implementation Guide for GYB Connect

## Overview

This guide covers the implementation of PCI DSS Requirements 7 (Restrict Access by Business Need-to-Know) and 8 (Identify Users and Authenticate Access) for the GYB Connect infrastructure. These requirements focus on implementing strong access control and authentication mechanisms to protect cardholder data.

## Table of Contents

- [Requirement 7: Strong Access Control](#requirement-7-strong-access-control)
- [Requirement 8: User Identification and Authentication](#requirement-8-user-identification-and-authentication)
- [Technical Implementation](#technical-implementation)
- [AWS IAM Identity Center Setup](#aws-iam-identity-center-setup)
- [Access Review Procedures](#access-review-procedures)
- [Monitoring and Compliance](#monitoring-and-compliance)

## Requirement 7: Strong Access Control

### 7.1 Limit Access to System Components and Cardholder Data

Our implementation enforces the principle of least privilege through:

1. **Role-Based Access Control (RBAC)**
   - API Lambda Role: Only accesses specific S3 buckets and DynamoDB tables
   - Data Processing Role: Read-only S3, batch DynamoDB operations
   - Read-Only Role: For monitoring and audit purposes only
   - Compliance Auditor Role: Limited to security findings and logs

2. **Resource-Based Policies**
   - S3 bucket policies restrict access to specific IAM roles
   - KMS key policies limit encryption/decryption to authorized services
   - DynamoDB resource policies ensure table-level access control

### 7.2 Establish an Access Control System

Our IAM Stack implements:

1. **MFA Boundary Policy**: All roles have a permission boundary that denies actions without MFA
2. **IAM Access Analyzer**: Continuously monitors for excessive permissions
3. **Least Privilege Policies**: Each role has minimal permissions required for its function

#### Implementation Details

```go
// Example: API Lambda Role - Least Privilege
{
    "Effect": "Allow",
    "Action": [
        "s3:GetObject",
        "s3:PutObject",
        "s3:DeleteObject"
    ],
    "Resource": "arn:aws:s3:::prod-gyb-uploads/*"
}
```

### 7.2.4 Review User Access Rights

Implement quarterly access reviews using:

1. **IAM Access Analyzer**: Identifies unused permissions
2. **CloudTrail Analysis**: Review actual API usage patterns
3. **Automated Reports**: Generate access reports for review

## Requirement 8: User Identification and Authentication

### 8.1 Assign Unique IDs to Each Person

- Each user gets a unique IAM Identity Center account
- Service accounts use unique IAM roles
- No shared credentials allowed

### 8.2 Strong Authentication Controls

1. **Password Requirements** (enforced via IAM Identity Center):
   - Minimum 12 characters
   - Mix of uppercase, lowercase, numbers, and symbols
   - Password history: 4 generations
   - Maximum age: 90 days

2. **Account Lockout**:
   - 6 failed attempts triggers 30-minute lockout
   - Alerts sent to security team

### 8.3 Multi-Factor Authentication (MFA)

**Mandatory MFA for all access to the CDE:**

1. **Human Users**: Phishing-resistant MFA via IAM Identity Center
2. **Service Accounts**: Use IAM roles with temporary credentials
3. **Root Account**: Hardware MFA token required

### 8.4 MFA Implementation

Our MFA Boundary Policy enforces MFA for all privileged actions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Deny",
            "NotAction": [
                "iam:CreateVirtualMFADevice",
                "iam:EnableMFADevice",
                "sts:GetSessionToken"
            ],
            "Resource": "*",
            "Condition": {
                "BoolIfExists": {
                    "aws:MultiFactorAuthPresent": "false"
                }
            }
        }
    ]
}
```

## Technical Implementation

### IAM Stack Components

1. **MFA Boundary Policy**: Applied to all roles to enforce MFA
2. **Service Roles**:
   - `prod-gyb-api-lambda-role`: API operations
   - `prod-gyb-data-processing-role`: Batch processing
   - `prod-gyb-readonly-role`: Monitoring access
   - `prod-gyb-compliance-auditor-role`: Audit access

3. **IAM Access Analyzer**: `prod-gyb-access-analyzer`

### Deployment

The IAM stack is deployed as part of the CDK infrastructure:

```bash
cdk deploy GybConnect-IAMStack
```

## AWS IAM Identity Center Setup

### Prerequisites

1. AWS Organization enabled
2. Admin access to the management account
3. Email addresses for all users

### Step-by-Step Setup

#### 1. Enable IAM Identity Center

```bash
# Navigate to IAM Identity Center in AWS Console
# Click "Enable" in your preferred region
```

#### 2. Configure Identity Source

For initial setup, use the Identity Center directory:

1. Go to Settings → Identity source
2. Select "Identity Center directory"
3. For production, consider external IdP (Okta, AD)

#### 3. Configure MFA Settings

1. Navigate to Settings → MFA
2. Configure:
   - **Prompt for MFA**: Every time they sign in
   - **MFA types**:
     - ✓ Authenticator apps
     - ✓ Security keys and biometric devices
     - ✗ SMS (not PCI DSS compliant)

#### 4. Create Permission Sets

Create the following permission sets matching our IAM roles:

##### Developer Permission Set

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": "sts:AssumeRole",
            "Resource": [
                "arn:aws:iam::*:role/dev-gyb-*",
                "arn:aws:iam::*:role/prod-gyb-readonly-role"
            ]
        }
    ]
}
```

##### Admin Permission Set

- Attach AWS managed policy: `AdministratorAccess`
- Session duration: 1 hour
- Require MFA for every session

##### Security Auditor Permission Set

```json
{
    "Effect": "Allow",
    "Action": "sts:AssumeRole",
    "Resource": "arn:aws:iam::*:role/*-gyb-compliance-auditor-role"
}
```

#### 5. Create Users and Groups

1. **Groups**:
   - `GYB-Developers`: Developer permission set
   - `GYB-Admins`: Admin permission set
   - `GYB-Auditors`: Security auditor permission set

2. **Users**:
   - Create individual users with corporate email
   - Assign to appropriate groups
   - Send invitation emails

#### 6. Assign Account Access

1. Go to AWS accounts → Select your accounts
2. Assign users/groups with appropriate permission sets
3. For production account:
   - Only `GYB-Admins` get write access
   - `GYB-Developers` get read-only access
   - `GYB-Auditors` get compliance auditor access

### User Onboarding Process

1. **Initial Setup**:

   ```
   1. User receives invitation email
   2. Sets up password meeting requirements
   3. Configures MFA (mandatory before first login)
   4. Downloads and configures AWS CLI v2
   ```

2. **CLI Configuration**:

   ```bash
   # Configure SSO
   aws configure sso
   # SSO start URL: https://your-domain.awsapps.com/start
   # SSO Region: us-west-1
   # Choose account and role
   
   # Login
   aws sso login --profile gyb-dev
   ```

## Access Review Procedures

### Quarterly Access Review Process

#### 1. Generate Access Reports

```bash
# Generate IAM credential report
aws iam generate-credential-report
aws iam get-credential-report --query 'Content' --output text | base64 -d > credential-report.csv

# List users with console access
aws iam list-users --query 'Users[?PasswordLastUsed!=`null`].[UserName,PasswordLastUsed]' --output table

# Check for unused roles
aws iam list-roles --query 'Roles[?AssumeRolePolicyDocument.Statement[0].Principal.Service==`lambda.amazonaws.com`].[RoleName,CreateDate]'
```

#### 2. Review IAM Access Analyzer Findings

```bash
# List all findings
aws accessanalyzer list-findings --analyzer-arn arn:aws:access-analyzer:us-west-1:123456789012:analyzer/prod-gyb-access-analyzer

# Review public and cross-account access
aws accessanalyzer list-findings --analyzer-arn $ANALYZER_ARN --filter '{"resourceType":{"eq":["AWS::S3::Bucket"]}}'
```

#### 3. Audit MFA Status

```bash
# Check MFA devices for all users
aws iam list-virtual-mfa-devices --query 'VirtualMFADevices[*].[User.UserName,SerialNumber,EnableDate]' --output table

# Identify users without MFA
aws iam generate-credential-report
aws iam get-credential-report --query 'Content' --output text | base64 -d | grep -E '^[^,]+,true,[^,]+,false'
```

#### 4. Document Review Results

Create a quarterly access review report including:

- Total number of active users
- Users without MFA (should be 0)
- Unused IAM roles or policies
- Access Analyzer findings and remediation
- Changes made since last review

### Automated Monitoring

#### CloudWatch Alarms for Access Violations

```python
# Example: Alert on root account usage
{
    "MetricName": "RootAccountUsage",
    "MetricNamespace": "CloudTrailMetrics",
    "MetricValue": "1",
    "MetricUnit": "Count"
}
```

#### EventBridge Rules for Real-time Alerts

1. **MFA Deactivation Alert**:

   ```json
   {
     "source": ["aws.iam"],
     "detail-type": ["AWS API Call via CloudTrail"],
     "detail": {
       "eventName": ["DeactivateMFADevice"]
     }
   }
   ```

2. **Privilege Escalation Detection**:

   ```json
   {
     "source": ["aws.iam"],
     "detail-type": ["AWS API Call via CloudTrail"],
     "detail": {
       "eventName": ["AttachUserPolicy", "PutUserPolicy", "AttachRolePolicy"]
     }
   }
   ```

## Monitoring and Compliance

### Key Metrics to Monitor

1. **Authentication Metrics**:
   - Failed login attempts
   - MFA bypass attempts
   - Unusual login locations or times

2. **Authorization Metrics**:
   - Access denied events
   - Privilege escalation attempts
   - Cross-account access patterns

3. **Compliance Metrics**:
   - Users with active MFA: Must be 100%
   - Password age compliance
   - Last activity for each user

### Security Hub Compliance Checks

Enable these PCI DSS controls in Security Hub:

- **[PCI.IAM.1]** IAM root user access key should not exist
- **[PCI.IAM.2]** IAM users should not have IAM policies attached
- **[PCI.IAM.3]** IAM policies should not allow full "*" administrative privileges
- **[PCI.IAM.4]** Hardware MFA should be enabled for the root user
- **[PCI.IAM.5]** Virtual MFA should be enabled for the root user
- **[PCI.IAM.6]** MFA should be enabled for all IAM users with console password
- **[PCI.IAM.7]** IAM user credentials unused for 90 days should be disabled
- **[PCI.IAM.8]** Password policies for IAM users should have strong configurations

### Integration with SIEM

Export CloudTrail logs to your SIEM for:

1. **Real-time alerting** on authentication anomalies
2. **Correlation** of access patterns with threat intelligence
3. **Compliance reporting** for audit requirements

## Best Practices

### For Administrators

1. **Never share AWS credentials**
2. **Use IAM roles instead of long-term access keys**
3. **Enable CloudTrail in all regions**
4. **Review Security Hub findings weekly**
5. **Conduct access reviews quarterly**

### For Developers

1. **Use temporary credentials from IAM roles**
2. **Never embed credentials in code**
3. **Request only the permissions you need**
4. **Use aws-vault or similar for local credential management**

### For Security Team

1. **Monitor all privileged actions**
2. **Investigate every root account usage**
3. **Automate compliance checks where possible**
4. **Maintain an incident response plan for credential compromise**

## Emergency Procedures

### Suspected Credential Compromise

1. **Immediate Actions**:

   ```bash
   # Disable the compromised user
   aws iam update-login-profile --user-name $USERNAME --no-password-reset-required
   
   # Delete access keys
   aws iam list-access-keys --user-name $USERNAME
   aws iam delete-access-key --user-name $USERNAME --access-key-id $KEY_ID
   
   # Revoke active sessions
   aws iam put-user-policy --user-name $USERNAME --policy-name DenyAll --policy-document '{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Action":"*","Resource":"*"}]}'
   ```

2. **Investigation**:
   - Review CloudTrail logs for unauthorized actions
   - Check for new resources created
   - Verify no data exfiltration occurred

3. **Recovery**:
   - Reset user credentials
   - Re-enable MFA
   - Review and update permissions if needed

## Conclusion

Implementing PCI DSS Requirements 7 and 8 provides strong access control and authentication for the GYB Connect infrastructure. Regular reviews, continuous monitoring, and strict enforcement of MFA ensure that only authorized users can access cardholder data environments.

Remember: Security is not a one-time setup but an ongoing process. Stay vigilant, keep systems updated, and regularly review access controls to maintain compliance.
