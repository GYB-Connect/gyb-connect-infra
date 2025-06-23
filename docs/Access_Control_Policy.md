# **GYB Connect Access Control Policy**

**Document Version**: 1.0  
**Effective Date**: January 2024  
**Review Date**: January 2025  
**Classification**: Internal Use Only

## **1. Purpose and Scope**

This Access Control Policy establishes the requirements for managing access to GYB Connect's information systems, applications, and data, with particular focus on protecting cardholder data and the Cardholder Data Environment (CDE).

**Scope**: All users, systems, applications, and data within the GYB Connect environment, with enhanced controls for the CDE.

## **2. Access Control Principles**

### **2.1 Principle of Least Privilege**

Users and systems are granted the minimum level of access necessary to perform their job functions. Access rights are restricted based on:

- Job role and responsibilities
- Business need-to-know
- Risk level of resources accessed
- Time-based limitations where appropriate

### **2.2 Defense in Depth**

Multiple layers of access controls protect sensitive resources:

- Network-level controls (VPC, Security Groups, NACLs)
- System-level controls (IAM roles, policies)
- Application-level controls (authentication, authorization)
- Data-level controls (encryption, field-level permissions)

### **2.3 Zero Trust Architecture**

No implicit trust is granted based on network location or user identity:

- Continuous verification of user and device identity
- All access requests are authenticated and authorized
- Micro-segmentation limits access scope
- Regular validation of access privileges

## **3. User Access Management**

### **3.1 Account Management**

**Account Creation**

- Formal request and approval process required
- Manager approval for all new accounts
- Security team approval for privileged accounts
- Automated provisioning where possible

**Account Types**

- **Standard User**: Basic business applications access
- **Privileged User**: Administrative or sensitive system access
- **Service Account**: Automated system-to-system access
- **Emergency Access**: Break-glass access for critical situations

**Account Lifecycle**

- Provisioning: Based on role templates and job functions
- Modification: Change requests with appropriate approval
- Regular Review: Quarterly access reviews for all accounts
- Deprovisioning: Immediate upon termination or role change

### **3.2 Authentication Requirements**

**Password Policy**

- Minimum 12 characters length
- Complexity requirements: uppercase, lowercase, numbers, symbols
- No reuse of last 4 passwords
- Maximum age: 90 days for privileged accounts, 180 days for standard
- Account lockout: 6 failed attempts, 30-minute lockout

**Multi-Factor Authentication (MFA)**

- **Required for all CDE access** (PCI DSS Requirement)
- **Required for all privileged accounts**
- **Required for remote access**
- Phishing-resistant MFA preferred (FIDO2, smart cards)
- SMS/voice backup methods where hardware tokens unavailable

**Single Sign-On (SSO)**

- AWS IAM Identity Center for all AWS access
- SAML integration for business applications
- Centralized authentication and session management
- Session timeout: 8 hours for standard users, 2 hours for privileged

### **3.3 Authorization Framework**

**Role-Based Access Control (RBAC)**

- Standardized roles aligned with job functions
- Role inheritance and delegation capabilities
- Segregation of duties enforcement
- Regular role definition reviews

**Attribute-Based Access Control (ABAC)**

- Fine-grained permissions based on attributes
- Dynamic access decisions based on context
- Time, location, and device-based restrictions
- Risk-adaptive access controls

**Standard User Roles**

- **Developer**: Development environment access, code repositories
- **Operations**: Production monitoring, limited administrative access
- **Analyst**: Business intelligence, reporting systems access
- **Support**: Customer service applications, limited data access

**Privileged Roles**

- **System Administrator**: Full administrative access to infrastructure
- **Database Administrator**: Database management and access
- **Security Administrator**: Security tool configuration and monitoring
- **Audit Administrator**: Read-only access for compliance verification

## **4. System Access Controls**

### **4.1 AWS Identity and Access Management (IAM)**

**IAM Policies**

- Least privilege principle enforcement
- Explicit deny where required
- Condition-based access controls
- Regular policy reviews and optimization

**IAM Roles**

- Service-specific roles for AWS resources
- Cross-account roles for multi-account architecture
- Temporary credentials for enhanced security
- Role chaining limitations

**Resource-Based Policies**

- S3 bucket policies for data access control
- KMS key policies for encryption key usage
- Lambda function resource policies
- API Gateway resource policies

### **4.2 Network Access Controls**

**Virtual Private Cloud (VPC)**

- Network segmentation and isolation
- Private subnets for sensitive resources
- Public subnet restrictions
- VPC endpoints for AWS service access

**Security Groups**

- Stateful firewall rules
- Principle of least privilege for network access
- Application-specific security groups
- Source-based access restrictions

**Network Access Control Lists (NACLs)**

- Stateless subnet-level filtering
- Explicit deny rules
- Defense in depth network security
- Egress filtering for data loss prevention

### **4.3 Application Access Controls**

**API Authentication**

- OAuth 2.0 / OpenID Connect
- API keys for service-to-service communication
- Rate limiting and throttling
- Request signing for critical APIs

**Database Access**

- Database-native authentication
- Connection encryption (SSL/TLS)
- Query-level access controls
- Database activity monitoring

**File System Access**

- Encryption at rest for all data
- Access logging and monitoring
- Regular permission audits
- Data loss prevention controls

## **5. Privileged Access Management**

### **5.1 Administrative Access**

**Just-in-Time (JIT) Access**

- Temporary elevated privileges
- Time-boxed access grants
- Approval workflow for privilege escalation
- Automatic privilege revocation

**Privileged Session Management**

- Session recording for audit trails
- Concurrent session limitations
- Privileged workstation requirements
- Network isolation for administrative tasks

**Emergency Access**

- Break-glass procedures for critical situations
- Emergency account activation process
- Comprehensive logging and monitoring
- Post-incident access review

### **5.2 Service Account Management**

**Service Account Standards**

- Dedicated accounts for each service
- No shared service credentials
- Regular credential rotation
- Least privilege principle application

**API Key Management**

- Secure key generation and storage
- Regular key rotation schedule
- Usage monitoring and anomaly detection
- Revocation procedures

**Certificate-Based Authentication**

- X.509 certificates for service authentication
- Certificate lifecycle management
- Certificate transparency and monitoring
- Automated certificate renewal

## **6. Remote Access**

### **6.1 Remote Access Requirements**

**VPN Access**

- Corporate VPN for all remote access
- Multi-factor authentication required
- Split tunneling prohibited for CDE access
- Device compliance verification

**Device Management**

- Managed device requirement for CDE access
- Device encryption and security controls
- Regular security patch deployment
- Remote wipe capabilities

**Network Restrictions**

- IP address allowlisting where feasible
- Geographic access restrictions
- Time-based access controls
- Bandwidth and connection monitoring

### **6.2 Cloud Access Security**

**Cloud Access Security Broker (CASB)**

- Data loss prevention policies
- Shadow IT discovery and control
- Risk scoring and assessment
- Compliance monitoring

**Zero Trust Network Access (ZTNA)**

- Application-specific access
- Device trust verification
- User behavior analytics
- Continuous access validation

## **7. Access Reviews and Monitoring**

### **7.1 Access Review Process**

**Quarterly Access Reviews**

- Manager approval of direct report access
- Role-based access validation
- Segregation of duties verification
- Unused account identification

**Annual Comprehensive Reviews**

- Complete access inventory
- Cross-functional access validation
- Privilege creep identification
- Policy compliance assessment

**Continuous Monitoring**

- Real-time access event monitoring
- Anomalous access pattern detection
- Privilege escalation alerts
- Failed access attempt tracking

### **7.2 Access Analytics**

**User Behavior Analytics (UBA)**

- Baseline user behavior establishment
- Anomaly detection and alerting
- Risk scoring and prioritization
- Machine learning-based analysis

**Access Intelligence**

- Access pattern analysis
- Unused permission identification
- Right-sizing recommendations
- Compliance gap identification

### **7.3 Audit and Compliance**

**Access Logging**

- Comprehensive access event logging
- Centralized log collection and storage
- Log integrity and protection
- Real-time log analysis

**Compliance Reporting**

- PCI DSS access control evidence
- SOX segregation of duties compliance
- Regulatory audit support
- Executive access dashboards

## **8. CDE-Specific Access Controls**

### **8.1 Cardholder Data Environment Access**

**CDE Access Requirements**

- Business justification required
- MFA mandatory for all access
- Privileged access monitoring
- Regular access recertification

**Network Segmentation**

- Isolated network segments for CDE
- Firewall controls between segments
- Network access control enforcement
- Regular segmentation validation

**Data Access Controls**

- Field-level access restrictions
- Data masking for non-production environments
- Query monitoring and analysis
- Data export controls

### **8.2 PCI DSS Compliance**

**Requirement 7 Implementation**

- Role-based access restrictions
- Need-to-know principle enforcement
- System access controls
- Regular access reviews

**Requirement 8 Implementation**

- Unique user identification
- Strong authentication controls
- Multi-factor authentication
- Secure credential management

## **9. Third-Party and Vendor Access**

### **9.1 Vendor Access Management**

**Access Authorization**

- Formal vendor access agreements
- Business justification documentation
- Security assessment completion
- Time-limited access grants

**Access Controls**

- Dedicated vendor accounts
- Restricted access scope
- Session monitoring and recording
- Regular access reviews

**Vendor Compliance**

- Security control verification
- PCI DSS compliance validation
- Regular security assessments
- Incident notification requirements

### **9.2 Business Partner Access**

**Partner Onboarding**

- Security questionnaire completion
- Access agreement execution
- Technical integration review
- Risk assessment completion

**Access Management**

- Federated identity integration
- Just-in-time access provisioning
- Attribute-based access controls
- Regular access validation

## **10. Training and Awareness**

### **10.1 Access Control Training**

**General User Training**

- Annual security awareness training
- Password security best practices
- MFA setup and usage
- Social engineering awareness

**Privileged User Training**

- Enhanced security training requirements
- Administrative best practices
- Incident response procedures
- Compliance responsibilities

**Security Team Training**

- Access control tool training
- Threat detection and response
- Compliance framework training
- Industry best practices

### **10.2 Security Awareness**

**Ongoing Communication**

- Monthly security newsletters
- Phishing simulation exercises
- Security incident lessons learned
- Policy update notifications

**Metrics and Reporting**

- Training completion tracking
- Phishing test results
- Security incident correlation
- Awareness program effectiveness

## **11. Incident Response**

### **11.1 Access-Related Incidents**

**Incident Types**

- Unauthorized access attempts
- Privilege escalation
- Account compromise
- Insider threats

**Response Procedures**

- Immediate access revocation
- Forensic evidence preservation
- Impact assessment
- Stakeholder notification

### **11.2 Recovery and Remediation**

**Account Recovery**

- Identity verification procedures
- Secure account reset process
- Access re-provisioning
- Security control validation

**System Hardening**

- Additional security controls
- Access pattern analysis
- Policy updates
- Control effectiveness testing

## **12. Technology and Tools**

### **12.1 Access Management Platforms**

**AWS IAM Identity Center**

- Centralized identity provider
- SAML federation capabilities
- Permission set management
- Multi-account access control

**Privileged Access Management (PAM)**

- Credential vaulting and rotation
- Session recording and monitoring
- Workflow-based access approval
- Risk-based access decisions

### **12.2 Monitoring and Analytics**

**Security Information and Event Management (SIEM)**

- Real-time event correlation
- Threat detection and alerting
- Compliance reporting
- Forensic analysis capabilities

**Identity Analytics**

- Access pattern analysis
- Risk scoring algorithms
- Peer group analysis
- Recommendation engines

## **13. Policy Enforcement**

### **13.1 Compliance Monitoring**

**Automated Controls**

- Policy violation detection
- Real-time alerting
- Automated remediation
- Compliance dashboard reporting

**Manual Reviews**

- Quarterly access assessments
- Annual policy compliance audits
- Spot checks and investigations
- Exception approval tracking

### **13.2 Enforcement Actions**

**Policy Violations**

- Warning and retraining
- Access restrictions
- Disciplinary action
- Account termination

**Security Incidents**

- Immediate access suspension
- Investigation procedures
- Legal action consideration
- Lessons learned integration

## **14. Related Documents**

- [Information Security Policy](./Information_Security_Policy.md)
- [Risk Management Policy](./Risk_Management_Policy.md)
- [Incident Response Plan](./Incident_Response_Plan.md)
- [Cryptography and Key Management Policy](./Cryptography_and_Key_Management_Policy.md)

---

**Document Owner**: Chief Technology Officer  
**Approved By**: [CTO Name]  
**Date**: [Approval Date]  
**Next Review**: January 2025

**For questions about this policy, contact**: <security@gybconnect.com>
