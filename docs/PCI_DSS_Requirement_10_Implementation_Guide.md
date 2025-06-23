# PCI DSS Requirement 10 Implementation Guide for GYB Connect

## Overview

This guide covers the implementation of PCI DSS Requirement 10 (Log and Monitor All Access to Network Resources and Cardholder Data) for the GYB Connect infrastructure. This requirement ensures comprehensive logging, centralized log management, real-time monitoring, and secure audit trail protection.

## Table of Contents

- [Requirement 10 Overview](#requirement-10-overview)
- [Technical Implementation](#technical-implementation)
- [Logging Stack Components](#logging-stack-components)
- [CloudTrail Configuration](#cloudtrail-configuration)
- [Real-Time Monitoring](#real-time-monitoring)
- [Log Retention and Protection](#log-retention-and-protection)
- [Alert Configuration](#alert-configuration)
- [Operational Procedures](#operational-procedures)
- [Compliance Verification](#compliance-verification)

## Requirement 10 Overview

PCI DSS Requirement 10 mandates logging and monitoring of all access to network resources and cardholder data. The sub-requirements include:

### 10.1 Audit Trail Implementation

- **10.1.1**: Audit trails must link all access to individual users
- **10.1.2**: Audit trails must be automated for all system components

### 10.2 Automated Audit Trails

- **10.2.1**: User access to cardholder data
- **10.2.2**: Administrative actions
- **10.2.3**: Access to audit trails
- **10.2.4**: Invalid logical access attempts
- **10.2.5**: Use of identification and authentication mechanisms
- **10.2.6**: Initialization of audit logs
- **10.2.7**: Creation and deletion of system-level objects

### 10.3 Record Specific Events

Record at minimum the following for each audit trail entry:

- User identification
- Type of event
- Date and time
- Success or failure indication
- Origination of event
- Identity of affected resource

### 10.4 Protect Audit Trail Files

- **10.4.1**: Protect audit trail files from unauthorized modifications
- **10.4.2**: Promptly back up audit trail files
- **10.4.3**: Write logs for external-facing technologies to secure centralized log server

### 10.5 Secure Centralized Time Synchronization

- **10.5.1**: Implement secure time synchronization technology

### 10.6 Review Logs and Security Events

- **10.6.1**: Review logs and security events for all system components
- **10.6.2**: Review logs of all critical system components regularly
- **10.6.3**: Follow up exceptions and anomalies

### 10.7 Retain Audit Trail History

- **10.7.1**: Retain audit trail history for at least one year
- **10.7.2**: At least three months immediately available for analysis

## Technical Implementation

### Architecture Overview

Our logging infrastructure implements a comprehensive solution:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CloudTrail    │    │  Application    │    │   EventBridge   │
│  (API Logging)  │    │     Logs        │    │   (Real-time)   │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────▼───────────────┐
                    │      CloudWatch Logs       │
                    │    (Central Processing)     │
                    └─────────────┬───────────────┘
                                  │
                    ┌─────────────▼───────────────┐
                    │    S3 Central Logging       │
                    │   (Long-term Storage)       │
                    │     Object Lock Enabled     │
                    └─────────────────────────────┘
```

### Core Components

1. **Logging Stack** (`stacks/logging_stack.go`)
2. **KMS Logging Key** (for encryption)
3. **S3 Central Logging Bucket** (immutable storage)
4. **CloudTrail** (API logging)
5. **CloudWatch Logs** (application logging)
6. **EventBridge Rules** (real-time alerting)
7. **SNS Topic** (alert distribution)

## Logging Stack Components

### 1. Central Logging S3 Bucket

**Purpose**: Immutable, long-term storage for all audit logs

**Features**:

- **Object Lock**: Compliance mode with 13-month retention
- **Encryption**: Customer-managed KMS key
- **Versioning**: Enabled for change tracking
- **Lifecycle**: Automatic archival to Glacier after 90 days
- **Access**: Blocked public access, restricted IAM policies

```go
// S3 bucket with PCI DSS compliant settings
ObjectLockEnabled: jsii.Bool(true),
Encryption: awss3.BucketEncryption_KMS,
EncryptionKey: props.LoggingKmsKey,
RemovalPolicy: awscdk.RemovalPolicy_RETAIN,
```

### 2. CloudTrail Configuration

**Purpose**: Comprehensive API call logging across all AWS services

**Configuration**:

- **Multi-region**: Captures events from all AWS regions
- **Global services**: Includes IAM, STS, CloudFront events
- **File validation**: Cryptographic integrity checking
- **Real-time delivery**: Streams to CloudWatch Logs

```go
IsMultiRegionTrail: jsii.Bool(true),
IncludeGlobalServiceEvents: jsii.Bool(true),
EnableFileValidation: jsii.Bool(true),
SendToCloudWatchLogs: jsii.Bool(true),
```

### 3. CloudWatch Log Groups

**Purpose**: Real-time log processing and metric extraction

**Features**:

- **Encryption**: Customer-managed KMS key
- **Retention**: 13 months minimum
- **Metric filters**: Convert log events to CloudWatch metrics
- **Real-time streams**: Immediate processing capability

### 4. SNS Security Alerts Topic

**Purpose**: Immediate notification of security events

**Configuration**:

- **Encryption**: Customer-managed KMS key
- **Email subscriptions**: Security team notifications
- **EventBridge integration**: Automated alert routing

## CloudTrail Configuration

### Logged Events

Our CloudTrail configuration captures:

1. **Management Events**:
   - API calls that modify AWS resources
   - Console sign-in events
   - IAM policy changes

2. **Global Service Events**:
   - IAM actions
   - STS token requests
   - CloudFront distributions

3. **File Validation**:
   - Cryptographic hashing
   - Tamper detection
   - Integrity verification

### Log Delivery

**Primary Storage**: S3 Central Logging Bucket

- Path: `s3://env-gyb-central-logs/cloudtrail-logs/`
- Format: JSON compressed with gzip
- Delivery: Within 15 minutes of API call

**Real-time Processing**: CloudWatch Logs

- Log Group: `/gyb-connect/env/application`
- Stream: Real-time delivery
- Processing: Metric filters and alarms

## Real-Time Monitoring

### EventBridge Rules

We monitor critical security events in real-time:

#### 1. IAM Policy Changes

```json
{
  "source": ["aws.iam"],
  "detail-type": ["AWS API Call via CloudTrail"],
  "detail": {
    "eventName": [
      "AttachUserPolicy", "DetachUserPolicy",
      "AttachRolePolicy", "DetachRolePolicy",
      "CreateRole", "DeleteRole"
    ]
  }
}
```

#### 2. S3 Bucket Policy Changes

```json
{
  "source": ["aws.s3"],
  "detail-type": ["AWS API Call via CloudTrail"],
  "detail": {
    "eventName": [
      "PutBucketPolicy", "DeleteBucketPolicy",
      "PutBucketAcl", "PutBucketPublicAccessBlock"
    ]
  }
}
```

#### 3. KMS Key Policy Changes

```json
{
  "source": ["aws.kms"],
  "detail-type": ["AWS API Call via CloudTrail"],
  "detail": {
    "eventName": [
      "CreateKey", "DisableKey",
      "ScheduleKeyDeletion", "PutKeyPolicy"
    ]
  }
}
```

### Metric Filters

CloudWatch metric filters convert log events to metrics:

#### Root Account Usage

```json
{
  "filterPattern": "{ ($.userIdentity.type = \"Root\") && ($.userIdentity.invokedBy NOT EXISTS) && ($.eventType != \"AwsServiceEvent\") }",
  "metricName": "RootAccountUsageCount",
  "metricValue": "1"
}
```

#### Failed Console Logins

```json
{
  "filterPattern": "{ ($.eventName = ConsoleLogin) && ($.responseElements.ConsoleLogin = \"Failure\") }",
  "metricName": "ConsoleLoginFailures",
  "metricValue": "1"
}
```

#### CloudTrail Configuration Changes

```json
{
  "filterPattern": "{ ($.eventName = CreateTrail) || ($.eventName = UpdateTrail) || ($.eventName = DeleteTrail) || ($.eventName = StartLogging) || ($.eventName = StopLogging) }",
  "metricName": "CloudTrailChanges",
  "metricValue": "1"
}
```

## Log Retention and Protection

### Retention Policies

**Short-term (Immediate Analysis)**:

- **CloudWatch Logs**: 13 months
- **Purpose**: Real-time monitoring and alerting
- **Access**: Operations team via IAM roles

**Long-term (Compliance Storage)**:

- **S3 Standard**: 90 days
- **S3 Glacier**: 90 days to 1 year
- **S3 Deep Archive**: 1+ years
- **Purpose**: Compliance and forensic analysis

### Protection Mechanisms

**Immutability**:

- S3 Object Lock in Compliance mode
- Cannot be deleted or modified during retention period
- Protects against accidental or malicious changes

**Encryption**:

- Customer-managed KMS key
- Separate key for logging workloads
- Automatic key rotation enabled

**Access Control**:

- IAM policies with least privilege
- Separate roles for read vs. admin access
- CloudTrail logs all access to audit logs

**Backup and Redundancy**:

- S3 cross-region replication (optional)
- Versioning enabled
- Multi-AZ storage by default

## Alert Configuration

### Critical Alerts (Immediate Response Required)

1. **Root Account Usage**
   - **Trigger**: Any root account activity
   - **Response Time**: Immediate (< 5 minutes)
   - **Action**: Investigate and document

2. **Multiple Failed Logins**
   - **Trigger**: 5+ failed attempts in 5 minutes
   - **Response Time**: Within 15 minutes
   - **Action**: Check for brute force attack

3. **CloudTrail Configuration Changes**
   - **Trigger**: Any modification to logging configuration
   - **Response Time**: Immediate
   - **Action**: Verify authorized change

### High Priority Alerts

4. **IAM Policy Changes**
   - **Trigger**: Changes to user/role policies
   - **Response Time**: Within 1 hour
   - **Action**: Review change authorization

5. **S3 Bucket Policy Changes**
   - **Trigger**: Changes to bucket access policies
   - **Response Time**: Within 1 hour
   - **Action**: Verify data remains secure

6. **KMS Key Policy Changes**
   - **Trigger**: Changes to encryption key policies
   - **Response Time**: Within 1 hour
   - **Action**: Verify encryption integrity

### Alert Routing

**SNS Topic**: `env-gyb-security-alerts`

- **Primary**: <security@gybconnect.com>
- **Secondary**: <devops@gybconnect.com>
- **Escalation**: CTO after 2 hours unacknowledged

**Message Format**:

```
Subject: [PCI DSS Alert] {Severity} - {Event Type}
Body: 
- Event: {EventName}
- User: {UserIdentity}
- Time: {EventTime}
- Source: {SourceIPAddress}
- Resources: {Resources}
- Response Required: {ResponseTime}
```

## Operational Procedures

### Daily Operations

#### Log Review Checklist

- [ ] Check CloudWatch dashboard for anomalies
- [ ] Review high-priority alerts from last 24 hours
- [ ] Verify CloudTrail is actively logging
- [ ] Confirm S3 log delivery is functioning
- [ ] Check for any failed metric filter deliveries

#### Weekly Operations

- [ ] Review all security alerts and responses
- [ ] Analyze failed login patterns
- [ ] Check log storage costs and optimization
- [ ] Verify log retention compliance
- [ ] Test alert mechanisms

#### Monthly Operations

- [ ] Review and update alerting thresholds
- [ ] Analyze log access patterns
- [ ] Test log restoration procedures
- [ ] Review IAM permissions for log access
- [ ] Update incident response procedures

### Incident Response Procedures

#### Step 1: Initial Assessment (Within 5 minutes)

1. **Acknowledge alert** and begin investigation
2. **Classify severity**:
   - Critical: Root access, configuration changes
   - High: Multiple failed logins, policy changes
   - Medium: Single failed login, normal operations
3. **Gather initial context** from CloudTrail logs

#### Step 2: Investigation (Within 15 minutes)

1. **Review complete event context**:

   ```bash
   # Example CloudTrail query
   aws logs filter-log-events \
     --log-group-name "/gyb-connect/prod/application" \
     --start-time 1640995200000 \
     --filter-pattern "{ $.eventName = \"ConsoleLogin\" }"
   ```

2. **Check related events** in same timeframe
3. **Verify user/source legitimacy**
4. **Document findings** in incident tracking system

#### Step 3: Response (Within 30 minutes)

1. **Contain threat** if malicious activity confirmed
2. **Notify stakeholders** based on severity
3. **Implement temporary controls** if needed
4. **Continue monitoring** for related activity

#### Step 4: Resolution and Learning

1. **Document complete timeline** of events
2. **Update detection rules** if gaps identified
3. **Improve alerting** based on lessons learned
4. **Conduct post-incident review**

### Log Analysis Commands

#### Common CloudTrail Queries

**Recent Root Account Activity**:

```bash
aws logs filter-log-events \
  --log-group-name "/gyb-connect/prod/application" \
  --filter-pattern "{ $.userIdentity.type = \"Root\" }" \
  --start-time $(date -d "1 hour ago" +%s)000
```

**Failed Login Attempts**:

```bash
aws logs filter-log-events \
  --log-group-name "/gyb-connect/prod/application" \
  --filter-pattern "{ $.eventName = \"ConsoleLogin\" && $.responseElements.ConsoleLogin = \"Failure\" }" \
  --start-time $(date -d "24 hours ago" +%s)000
```

**IAM Policy Changes**:

```bash
aws logs filter-log-events \
  --log-group-name "/gyb-connect/prod/application" \
  --filter-pattern "{ $.eventSource = \"iam.amazonaws.com\" && $.eventName = \"AttachUserPolicy\" }" \
  --start-time $(date -d "7 days ago" +%s)000
```

**S3 Access from External IPs**:

```bash
aws logs filter-log-events \
  --log-group-name "/gyb-connect/prod/application" \
  --filter-pattern "{ $.eventSource = \"s3.amazonaws.com\" && $.sourceIPAddress != \"10.*\" }" \
  --start-time $(date -d "1 hour ago" +%s)000
```

## Compliance Verification

### PCI DSS Requirement Mapping

| Requirement | Implementation | Verification Method |
|-------------|----------------|-------------------|
| 10.1.1 | CloudTrail user identification | Audit log samples |
| 10.1.2 | Automated CloudTrail logging | Configuration review |
| 10.2.1 | Data access logging | Application log analysis |
| 10.2.2 | Admin action logging | CloudTrail policy changes |
| 10.2.3 | Audit trail access logging | S3 access logs |
| 10.2.4 | Failed access logging | Failed login metrics |
| 10.2.5 | Auth mechanism logging | IAM CloudTrail events |
| 10.2.6 | Log initialization | CloudTrail start events |
| 10.2.7 | Object creation/deletion | S3 and IAM CloudTrail |
| 10.3.x | Required data elements | Log format verification |
| 10.4.1 | Unauthorized modification protection | Object Lock verification |
| 10.4.2 | Audit trail backup | S3 replication status |
| 10.4.3 | Centralized logging | Log delivery verification |
| 10.5.1 | Time synchronization | NTP configuration check |
| 10.6.1 | Log review processes | Review documentation |
| 10.6.2 | Critical component review | Alert configuration |
| 10.6.3 | Exception follow-up | Incident response logs |
| 10.7.1 | One-year retention | S3 lifecycle policies |
| 10.7.2 | Three-month availability | CloudWatch retention |

### Audit Evidence Collection

#### 1. Configuration Evidence

- CloudTrail configuration screenshots
- S3 bucket policy documents
- KMS key policy documents
- IAM role and policy definitions

#### 2. Process Evidence

- Incident response procedures
- Log review checklists
- Alert escalation procedures
- Training records

#### 3. Technical Evidence

- Log samples showing required data elements
- Object Lock retention verification
- Encryption key usage reports
- Alert testing results

### Self-Assessment Checklist

**Requirement 10.1**: Audit Trail Implementation

- [ ] All API calls captured in CloudTrail
- [ ] User identification present in all logs
- [ ] Automated logging enabled for all components

**Requirement 10.2**: Automated Audit Trails

- [ ] User access to cardholder data logged
- [ ] Administrative actions captured
- [ ] Audit trail access monitored
- [ ] Failed access attempts recorded
- [ ] Authentication events logged
- [ ] Log initialization events captured
- [ ] System object changes tracked

**Requirement 10.3**: Required Data Elements

- [ ] User identification in logs
- [ ] Event type recorded
- [ ] Date and time stamps accurate
- [ ] Success/failure indication present
- [ ] Event origination captured
- [ ] Affected resource identification

**Requirement 10.4**: Audit Trail Protection

- [ ] Logs protected from unauthorized modification (Object Lock)
- [ ] Regular backups performed (S3 durability)
- [ ] Centralized logging implemented (S3 bucket)

**Requirement 10.5**: Time Synchronization

- [ ] NTP configured on all systems
- [ ] Time synchronization verified
- [ ] Time zone consistency maintained

**Requirement 10.6**: Log Review

- [ ] Daily log review procedures documented
- [ ] Critical components monitored
- [ ] Exception handling processes defined
- [ ] Alert response procedures implemented

**Requirement 10.7**: Retention Requirements

- [ ] One-year retention implemented (S3 lifecycle)
- [ ] Three-month immediate availability (CloudWatch)
- [ ] Archive procedures documented

## Troubleshooting

### Common Issues and Solutions

#### CloudTrail Not Logging

**Symptoms**: No events appearing in CloudWatch Logs
**Causes**:

- IAM permissions insufficient
- CloudWatch Logs destination misconfigured
- Region mismatch

**Resolution**:

```bash
# Check CloudTrail status
aws cloudtrail get-trail-status --name prod-gyb-cloudtrail

# Verify CloudWatch Logs delivery
aws logs describe-log-groups --log-group-name-prefix "/gyb-connect"

# Check IAM role permissions
aws iam get-role-policy --role-name CloudTrail_CloudWatchLogs_Role --policy-name CloudWatchLogsPolicy
```

#### S3 Object Lock Issues

**Symptoms**: Cannot delete old log files
**Causes**:

- Object Lock retention not expired
- Incorrect IAM permissions
- Governance mode vs. Compliance mode

**Resolution**:

```bash
# Check Object Lock status
aws s3api get-object-lock-configuration --bucket prod-gyb-central-logs

# Verify retention settings
aws s3api get-object-retention --bucket prod-gyb-central-logs --key cloudtrail-logs/file.json

# Check bucket policy
aws s3api get-bucket-policy --bucket prod-gyb-central-logs
```

#### Alert Not Firing

**Symptoms**: No notifications for critical events
**Causes**:

- EventBridge rule misconfigured
- SNS topic permissions wrong
- Metric filter pattern incorrect

**Resolution**:

```bash
# Test EventBridge rule
aws events test-event-pattern \
  --event-pattern file://rule-pattern.json \
  --event file://test-event.json

# Verify SNS subscription
aws sns list-subscriptions-by-topic --topic-arn arn:aws:sns:region:account:prod-gyb-security-alerts

# Check CloudWatch metrics
aws cloudwatch get-metric-statistics \
  --namespace CloudTrailMetrics \
  --metric-name RootAccountUsageCount \
  --start-time 2024-01-01T00:00:00Z \
  --end-time 2024-01-02T00:00:00Z \
  --period 3600 \
  --statistics Sum
```

## Cost Optimization

### Cost Breakdown (Monthly Estimates)

**CloudTrail**: $2.00 per 100,000 events

- Typical usage: ~500,000 events/month
- Cost: ~$10/month

**CloudWatch Logs**: $0.50 per GB ingested

- Estimated ingestion: 20 GB/month
- Cost: ~$10/month

**S3 Storage**:

- Standard: $0.023 per GB (first 90 days)
- Glacier: $0.004 per GB (90 days - 1 year)
- Deep Archive: $0.00099 per GB (1+ years)
- Estimated monthly storage: 10 GB new + archive
- Cost: ~$15/month growing over time

**Data Transfer**: Minimal (within AWS)

- Cost: ~$1/month

**Total Estimated Cost**: ~$36/month

### Optimization Strategies

1. **Lifecycle Policies**: Automatic archival to cheaper storage
2. **Log Filtering**: Reduce unnecessary log volume
3. **Metric Filters**: Process only required events
4. **Regional Optimization**: Single region for development
5. **Alert Tuning**: Reduce false positive notifications

## Conclusion

This implementation provides comprehensive logging and monitoring capabilities that exceed PCI DSS Requirement 10 mandates. The centralized, immutable, and encrypted logging infrastructure ensures full audit capability while maintaining cost efficiency and operational simplicity.

Key benefits:

- **Complete audit trail** of all API and administrative actions
- **Real-time security alerting** for immediate threat response
- **Immutable log storage** preventing tampering or accidental deletion
- **Cost-effective archival** with intelligent lifecycle management
- **Automated compliance** reducing manual review burden

Regular review and updates of this implementation ensure continued compliance and security effectiveness.
