# **PCI DSS SAQ D-SP Compliance Roadmap for GYB Connect**

## **1\. Introduction**

This document provides a detailed analysis of our current GYB Connect AWS infrastructure (as defined in our CDK project) and outlines a strategic roadmap for us to achieve compliance with the **Payment Card Industry Data Security Standard (PCI DSS) v4.0**, specifically for the **SAQ D for Service Providers (SAQ D-SP)**.

Our current architecture, built with the AWS CDK in Golang, establishes a solid foundation with excellent practices like modular stacks and environment-specific deployments. This Infrastructure as Code (IaC) approach is crucial for us to create repeatable, auditable, and secure environments.

The **goal** of this roadmap is to evolve our existing infrastructure into a secure, compliant **Cardholder Data Environment (CDE)**. The CDE includes all people, processes, and technology that we use to store, process, or transmit cardholder data, plus any systems that are connected to or could impact the security of that environment. As a service provider, the SAQ D-SP standard requires **the most rigorous set of controls,** which we will address systematically.

This is a living document that we should update as our architecture evolves.

## **2\. High-Level Architectural Recommendations**

Before diving into the specific requirements, here are three overarching strategic changes that will provide the foundation for all other controls we need to implement:

1. **Adopt a Multi-Account Strategy with AWS Organizations**: While our current setup separates dev and prod logically, PCI DSS requires **strict separation**. The industry best practice, and our most secure approach, is to use separate AWS accounts for our development, staging, and production environments. This provides the strongest security and billing boundary. We will use **AWS Organizations** to manage these accounts and apply **Service Control Policies (SCPs)** to enforce global security guardrails (e.g., prevent services from being launched in unapproved regions).  
2. **Establish a Dedicated "Security" AWS Account**: We will create a separate AWS account dedicated to security services. This account will be the central hub for all our logs (CloudTrail, VPC Flow Logs, etc.), security findings (GuardDuty, Security Hub, Inspector), and will host our security tooling. Access to this account must be **extremely limited.**  
3. **Implement Centralized Identity Management**: We will use **AWS IAM Identity Center (formerly AWS SSO)** as the single entry point for all human access to our AWS accounts. This will allow us to enforce **Multi-Factor Authentication (MFA)** universally (a critical SAQ D requirement) and manage permissions from one place, adhering to the principle of least privilege.

## **3\. PCI DSS Requirements: Analysis and Action Plan**

Each section below maps directly to a PCI DSS goal. We will analyze our current infrastructure against each requirement and define a clear, actionable plan for us to achieve compliance.

### **Requirement 1: Install and Maintain Network Security Controls**

*Our VpcStack is a great start for production, but we need to harden it significantly and use it consistently. We must treat the CDE as a fortress, with **every entry and exit point strictly controlled.***

#### **Current State Analysis:**

* **Production**: We have correctly defined a custom VPC with public and private subnets in vpc\_stack.go.  
* **Development**: Our rds\_stack.go uses the default VPC and places the database in public subnets, which is a major security risk and not suitable even for development in a PCI context.  
* **Security Groups**: We have a basic security group for the RDS instance, allowing traffic from within the VPC. This is a good start, but it needs to be more granular.  
* **Missing Controls**: We have not yet defined any Network ACLs, VPC Flow Logs, or AWS WAF.

#### **✅ Required Changes & Next Steps:**

1. **Enforce Strict CDE Segmentation (Req. 1.2, 1.3)**:  
   * **Action**: We will modify our CDK deployment logic (gyb\_connect.go) to **always** use the dedicated VpcStack for any environment that will become part of the CDE, including staging. **We will not use the default VPC for any in-scope resources.**  
   * **Action**: We will create a separate, more isolated VPC for our production CDE. Within this VPC, we will create more granular subnets. For example:  
     * public-subnets: For our load balancers and NAT Gateways.  
     * app-subnets (Private): For our AppRunner services and Lambda functions.  
     * data-subnets (Private & Isolated): For our RDS and DynamoDB (if accessed via VPC endpoint). These subnets will have no route to the internet.  
2. **Implement Stateful and Stateless Firewalls (Req. 1.2.1, 1.3.1)**:  
   * **Action**: We will enhance our Security Groups. Instead of allowing traffic from the entire VPC CIDR, we will create specific SGs for our AppRunner services, Lambdas, and RDS instance. Rules will only allow traffic from the specific source SG on the specific port required (e.g., the AppRunner SG can talk to the RDS SG on port 5432).  
   * **Action**: We will add **Network Access Control Lists (NACLs)** to our VpcStack. NACLs act as a stateless firewall at the subnet level, providing a second layer of defense. We will configure them with explicit deny-all rules, only allowing traffic required by our application.  
3. **Enable Comprehensive Network Logging (Req. 10.2)**:  
   * **Action**: We will enable **VPC Flow Logs** on our production VPC. We will configure the logs to be delivered to a centralized, write-once S3 bucket in our dedicated Security Account. This is critical for our incident response and forensic analysis capabilities.  
4. **Secure the Network Perimeter (Req. 1.4, 6.4.2)**:  
   * **Action**: We will implement **AWS WAF** and associate it with our public-facing entry points. We will start with AWS Managed Rule Groups like AWSManagedRulesCommonRuleSet and AWSManagedRulesSQLiRuleSet, deploying them in 'Count' mode first to monitor for false positives before switching to 'Block' mode.  
   * **Action**: We will use **AWS Shield Standard** (enabled by default) and consider **Shield Advanced** for robust DDoS protection.  
5. **Use VPC Endpoints (Req. 1.3.4)**:  
   * **Action**: To ensure our services within private subnets can access other AWS services without traversing the public internet, we will create **VPC Gateway and Interface Endpoints** in our VpcStack.

### **Requirement 2: Apply Secure Configurations**

*IaC is a huge win for us here, as it ensures repeatable, secure configurations. Now, we must harden those configurations and eliminate insecure defaults.*

#### **Current State Analysis:**

* Our use of CDK enforces configuration standards, which is excellent.  
* However, we have insecure defaults present (e.g., default VPC usage in dev, overly permissive CORS rules, AutoDeleteObjects enabled for S3).

#### **✅ Required Changes & Next Steps:**

1. **Harden All Service Configurations (Req. 2.2)**:  
   * **Action**: In our s3\_stack.go and apigateway\_stack.go, we will immediately remove wildcard (\*) CORS origins and replace them with the specific domains of our Amplify frontend for each environment.  
   * **Action**: In s3\_stack.go and rds\_stack.go, we will ensure the production environment settings for RemovalPolicy are RETAIN and DeletionProtection is true. Our deployment scripts must not be able to override this for production.  
   * **Action**: For AppRunner, Lambdas, and any other compute, we will ensure they run with the minimum necessary permissions. Their IAM execution roles will not have wildcard permissions.  
2. **Establish Secure Baselines (Req. 2.2.1)**:  
   * **Action**: We will document our CDK code as our configuration standard. For any manual configurations, we will create a formal hardening guide based on **CIS Benchmarks**.  
3. **Encrypt All Non-Console Administrative Access (Req. 2.2.7)**:  
   * **Action**: We will ensure all access to the AWS console and CLI/API requires MFA (see Req. 8). All management traffic is already encrypted via TLS by the AWS APIs, but we must ensure we are not using any insecure protocols for other management tasks.

### **Requirements 3 & 4: Protect Stored and Transmitted Account Data**

*Our current setup correctly enables encryption-at-rest for all data stores, which is a great start. Our next step is to gain full control over the encryption keys and ensure all data in transit is secure.*

#### **Current State Analysis:**

* Our S3, DynamoDB, and RDS stacks correctly enable server-side encryption.  
* However, we use AWS-managed keys, which offer less control and auditability than customer-managed keys.

#### **✅ Required Changes & Next Steps:**

1. **Implement Customer-Managed Keys (CMKs) (Req. 3.5, 3.6)**:  
   * **Action**: We will create a new KmsStack in our CDK project. In this stack, we will define dedicated **AWS KMS Customer-Managed Keys (CMKs)** for each data type and environment.  
   * **Action**: We will update our stacks to use these CMKs. This is a critical step because it gives us direct control over the key lifecycle and generates detailed audit trails in CloudTrail for every use of the key (Encrypt, Decrypt, GenerateDataKey), providing irrefutable evidence to auditors.  
   * **Action**: We will enable **automatic key rotation** (annual) for all our CMKs.  
   * **Action**: We will define strict KMS Key Policies that limit key usage to only the specific IAM roles of the services that need them.  
2. **Automate Data Discovery (Req. 3.1, 12.5.2)**:  
   * **Action**: We will enable **Amazon Macie** in our AWS account and configure it to continuously scan our S3 buckets to ensure no unencrypted or sensitive data is accidentally stored.  
3. **Enforce Encryption in Transit (Req. 4.1)**:  
   * **Action**: We will configure our Amplify frontend to only serve traffic over HTTPS using a modern TLS policy.  
   * **Action**: In our ApiGatewayStack, we will configure a custom domain and enforce a TLS 1.2+ security policy.  
   * **Action**: For our RDS instance, we will enforce encrypted connections by setting the rds.force\_ssl parameter in a custom parameter group.

### **Requirements 5, 6, & 11: Vulnerability Management, Secure Software, and Testing**

*This is a significant area for new implementation. We need to protect against malware, secure our software supply chain, and implement a regular testing cadence.*

#### **Current State Analysis:**

* We have not yet defined malware protection.  
* We have not yet integrated vulnerability scanning for code or infrastructure.  
* We do not have a formal testing (ASV, pentesting) process mentioned.

#### **✅ Required Changes & Next Steps:**

1. **Deploy Malware and Vulnerability Protection (Req. 5.2, 11.3)**:  
   * **Action**: We will enable **Amazon GuardDuty** across our entire AWS Organization to act as our Intrusion Detection System (IDS) and enable its **Malware Protection** feature.  
   * **Action**: We will enable **Amazon Inspector** to continuously scan our container images and Lambda functions for known software vulnerabilities (CVEs).  
   * **Action**: For any future EC2 instances, we must deploy an **anti-malware/EDR solution**.  
2. **Secure Our CI/CD Pipeline and Frontend (Req. 6.2, 6.4.3)**:  
   * **Action**: We will integrate static application security testing (**SAST**) and software composition analysis (**SCA**) tools into our CI/CD pipeline.  
   * **Action (Critical for Frontend)**: For our payment pages on Amplify, we must fulfill PCI DSS Req. 6.4.3 by creating a script inventory, implementing a strict **Content Security Policy (CSP)**, and using **Subresource Integrity (SRI)**.  
3. **Implement a Formal Testing and Change Management Program (Req. 11.4, 6.5)**:  
   * **Action**: We will schedule and perform **quarterly external vulnerability scans** by a certified Approved Scanning Vendor (ASV).  
   * **Action**: We will schedule and perform **annual internal and external penetration tests**.  
   * **Action**: We will formalize our change control process. All CDE changes must be documented, analyzed for security impact, and approved.

### **Requirements 7 & 8: Strong Access Control**

*This is another critical area for improvement. We must rigorously apply the principle of least privilege, and MFA is non-negotiable for SAQ D.*

#### **Current State Analysis:**

* We have not defined specific IAM roles for applications.  
* We have not defined MFA enforcement.  
* We do not have a centralized user management system in place.

#### **✅ Required Changes & Next Steps:**

1. **Enforce Universal MFA and SSO (Req. 8.4)**:  
   * **Action**: We will deploy **AWS IAM Identity Center** and configure it as the *only* way for our team to access the AWS Console and CLI.  
   * **Action**: We will enforce phishing-resistant **MFA** for every user in IAM Identity Center and for the root user of every account.  
2. **Implement Least Privilege for Services (Req. 7.2)**:  
   * **Action**: We will create a new IamStack and define granular IAM execution roles for our AppRunner service(s) and every Lambda function.  
   * **Rule of Thumb**: A role will only have the permissions it needs for its job.  
3. **Implement Periodic Access Reviews (Req. 7.2.4)**:  
   * **Action**: We will establish a formal, documented process to review all user and role access rights at least every six months, using **IAM Access Analyzer** to help identify unused permissions.

### **Requirement 10: Log and Monitor All Access**

*Logging is the foundation of our detection and forensics strategy. We must log all actions in the CDE, and those logs must be protected and actively monitored.*

#### **Current State Analysis:**

* Logging is not explicitly configured in our CDK project. Default service logs may exist but are not centralized, protected, or monitored.

#### **✅ Required Changes & Next Steps:**

1. **Centralize and Protect All Logs (Req. 10.2, 10.3, 10.5)**:  
   * **Action**: We will create a dedicated LoggingStack.  
   * **Action**: In this stack, we will create a central S3 bucket in our Security Account, configured with **Object Lock (in Compliance Mode)** to make logs immutable.  
   * **Action**: We will enable **AWS CloudTrail** for our entire organization and deliver all logs to this central bucket.  
   * **Action**: We will configure our services (AppRunner, API Gateway, WAF) to export their logs to CloudWatch Log Groups and stream them to the central S3 bucket.  
2. **Implement Real-Time Alerting (Req. 10.4, 10.7)**:  
   * **Action**: We will use **Amazon CloudWatch** to create alerts for critical security events found in our logs.  
   * **Examples of Critical Alerts**: Root user login, MFA deactivation, changes to IAM policies, disabling CloudTrail or GuardDuty, high-severity findings from GuardDuty or Inspector.  
   * **Action**: We will integrate these alerts with an **SNS topic** to notify our security team immediately.

### **Requirement 9 & 12: Physical Security and Security Policy**

*These requirements are less about code and more about process and documentation, but they are equally important for our compliance.*

#### **✅ Required Changes & Next Steps:**

1. **Leverage AWS Compliance (Req. 9\)**:  
   * **Action**: For physical security of our cloud infrastructure, we will rely on AWS. We will download the **AWS Attestation of Compliance (AOC)** and other reports from **AWS Artifact** as evidence for our auditors.  
2. **Develop Formal Security Policies (Req. 12.1)**:  
   * **Action**: We must support all technical controls with formal, written information security policies. We will create and maintain, at a minimum: an overall Information Security Policy, an Incident Response Plan, a Risk Management Policy, an Access Control Policy, a Cryptography and Key Management Policy, and a Secure Software Development Lifecycle (SDLC) Policy.

## **4\. Summary of High-Priority Next Steps**

To begin, we will focus on these foundational tasks, as they will have the most significant impact on our security posture and are prerequisites for many other controls:

1. **Isolate Environments**: We will immediately begin setting up separate AWS accounts for dev, staging, and prod using AWS Organizations and will stop using the default VPC for any resources.  
2. **Implement IAM Identity Center**: We will centralize all human access and enforce MFA for every user. This is a non-negotiable control for us.  
3. **Enable Foundational Security Services**: We will turn on AWS Security Hub, Amazon GuardDuty, and Amazon Inspector across all our accounts.  
4. **Harden Network Controls**: We will deploy AWS WAF, implement stricter Security Group and NACL rules, and enable VPC Flow Logs.  
5. **Centralize and Protect Logs**: We will create our immutable S3 logging bucket and configure CloudTrail to log everything to it.

By systematically implementing these changes, we will build a robust, secure, and compliant infrastructure capable of meeting the rigorous demands of PCI DSS SAQ D-SP.
