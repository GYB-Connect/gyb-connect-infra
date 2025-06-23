# **AWS Artifact Compliance Guide for GYB Connect**

**Document Version**: 1.0  
**Effective Date**: January 2024  
**Review Date**: January 2025  
**Classification**: Internal Use Only

## **1. Purpose and Scope**

This guide provides instructions for accessing AWS compliance documentation through AWS Artifact to satisfy PCI DSS Requirement 9 (Physical Security) for GYB Connect's cloud infrastructure. These documents provide evidence of AWS's physical security controls for auditors.

**Scope**: AWS Artifact document access for PCI DSS compliance evidence collection.

## **2. Understanding AWS Shared Responsibility Model**

### **2.1 AWS Responsibilities (Security OF the Cloud)**

**Physical Security**

- Data center physical security
- Hardware lifecycle management
- Network infrastructure protection
- Environmental controls and monitoring

**Infrastructure Security**

- Host operating system patching
- Hypervisor security
- Network configuration
- Service availability and resilience

### **2.2 Customer Responsibilities (Security IN the Cloud)**

**Data Protection**

- Data encryption configuration
- Access control implementation
- Network security configuration
- Operating system updates (for EC2)

**Application Security**

- Application code security
- Identity and access management
- Security group configuration
- SSL/TLS certificate management

## **3. AWS Artifact Overview**

### **3.1 What is AWS Artifact**

AWS Artifact is your central resource for compliance-related information that matters to your organization. It provides on-demand access to AWS security and compliance reports and select online agreements.

**Key Features**

- Self-service compliance report downloads
- AWS security certifications and attestations
- Compliance documentation repository
- Automated report delivery options

### **3.2 Available Reports**

**Security Certifications**

- SOC 1 Type II reports
- SOC 2 Type II reports
- SOC 3 reports
- ISO 27001/27017/27018 certificates

**Compliance Attestations**

- PCI DSS Attestation of Compliance (AOC)
- FIPS 140-2 compliance reports
- FedRAMP authorization documentation
- Industry-specific compliance reports

## **4. Accessing AWS Artifact**

### **4.1 Prerequisites**

**AWS Account Requirements**

- Valid AWS account with appropriate permissions
- IAM permissions for AWS Artifact access
- Business relationship agreement with AWS

**Required IAM Permissions**

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "artifact:Get*",
                "artifact:List*",
                "artifact:DownloadAgreement"
            ],
            "Resource": "*"
        }
    ]
}
```

### **4.2 Step-by-Step Access Guide**

**Step 1: Access AWS Artifact**

1. Log into the AWS Management Console
2. Navigate to AWS Artifact service
3. Or visit: <https://console.aws.amazon.com/artifact/>

**Step 2: Accept AWS Customer Agreement**

1. Review the AWS Customer Agreement
2. Accept terms and conditions
3. Provide company information as required

**Step 3: Browse Available Reports**

1. Navigate to "Reports" section
2. Filter by compliance standard (PCI DSS)
3. Review available report types and dates

**Step 4: Download Required Reports**

1. Select desired compliance reports
2. Review report descriptions and validity periods
3. Download reports in PDF format
4. Store securely for audit purposes

## **5. Required Documents for PCI DSS Compliance**

### **5.1 PCI DSS Specific Documents**

**AWS PCI DSS Attestation of Compliance (AOC)**

- **File**: AWS PCI DSS AOC Level 1 Service Provider
- **Validity**: Check for current year certification
- **Content**: AWS PCI DSS compliance attestation
- **Usage**: Provide to QSA as evidence of AWS compliance

**AWS PCI DSS Responsibility Matrix**

- **File**: AWS PCI DSS Shared Responsibility Matrix
- **Content**: Detailed breakdown of AWS vs. customer responsibilities
- **Usage**: Understanding compliance requirements

### **5.2 Supporting Security Documents**

**SOC 2 Type II Reports**

- **Content**: Detailed security control testing results
- **Focus Areas**: Security, availability, processing integrity
- **Usage**: Additional security control evidence

**ISO 27001 Certificate**

- **Content**: Information Security Management System certification
- **Scope**: AWS global infrastructure and services
- **Usage**: International security standard compliance evidence

**AWS Security Whitepaper**

- **Content**: Comprehensive security overview
- **Topics**: Physical security, network security, data protection
- **Usage**: Technical security implementation details

## **6. Document Management and Storage**

### **6.1 Document Security**

**Storage Requirements**

- Encrypted storage for all compliance documents
- Access control to compliance documentation
- Version control and document lifecycle management
- Backup and recovery procedures

**Access Controls**

- Limit access to authorized personnel only
- Audit trail for document access
- Time-limited access for external auditors
- Secure document sharing procedures

### **6.2 Document Organization**

**Recommended Folder Structure**

```
Compliance Documents/
├── AWS Artifact Reports/
│   ├── PCI DSS/
│   │   ├── 2024/
│   │   │   ├── AWS_PCI_DSS_AOC_2024.pdf
│   │   │   ├── AWS_PCI_DSS_Responsibility_Matrix_2024.pdf
│   │   │   └── AWS_PCI_DSS_Evidence_Package_2024.pdf
│   │   └── Previous Years/
│   ├── SOC Reports/
│   │   ├── SOC_2_Type_II_2024.pdf
│   │   └── Previous Reports/
│   └── ISO Certificates/
│       ├── ISO_27001_Certificate_2024.pdf
│       └── Previous Certificates/
├── GYB Connect Compliance/
│   ├── Internal Assessments/
│   ├── Third Party Audits/
│   └── Remediation Evidence/
└── Audit Evidence/
    ├── Current Audit/
    ├── Previous Audits/
    └── Ongoing Monitoring/
```

## **7. Using AWS Compliance Documentation**

### **7.1 For Internal Compliance**

**Self-Assessment Activities**

- Compare AWS controls to PCI DSS requirements
- Identify customer responsibility areas
- Plan implementation of required controls
- Document compliance gaps and remediation

**Risk Assessment**

- Review AWS security controls effectiveness
- Assess residual risks in shared responsibility
- Plan additional security measures
- Update risk register and treatment plans

### **7.2 For External Audits**

**Auditor Preparation**

- Share AWS AOC with Qualified Security Assessor (QSA)
- Provide responsibility matrix documentation
- Explain shared responsibility model
- Demonstrate customer-implemented controls

**Evidence Package**

- AWS compliance certifications
- Customer security implementations
- Monitoring and logging evidence
- Incident response capabilities

## **8. Keeping Documentation Current**

### **8.1 Regular Updates**

**Monthly Checks**

- Review AWS Artifact for new reports
- Check for updated compliance certificates
- Monitor expiration dates of current documents
- Update document inventory

**Quarterly Reviews**

- Compare new reports with existing documentation
- Update internal compliance assessments
- Review shared responsibility understanding
- Plan for upcoming audit requirements

### **8.2 Automated Monitoring**

**AWS Config Rules**

- Monitor AWS service compliance status
- Automated compliance reporting
- Configuration drift detection
- Compliance dashboard maintenance

**CloudWatch Alarms**

- Monitor for compliance-related events
- Alert on configuration changes
- Track security metric thresholds
- Integrate with incident response

## **9. Common AWS Artifact Issues and Solutions**

### **9.1 Access Issues**

**Problem**: Cannot access AWS Artifact
**Solution**:

- Verify IAM permissions are correctly configured
- Ensure AWS account is in good standing
- Check if Customer Agreement has been accepted

**Problem**: Reports not available for download
**Solution**:

- Verify you have necessary business agreements with AWS
- Check if specific compliance programs require enrollment
- Contact AWS support for enterprise-level access

### **9.2 Document Issues**

**Problem**: Reports are outdated or expired
**Solution**:

- Check AWS Artifact regularly for updated reports
- Understand certification renewal cycles
- Plan audit timing around report availability

**Problem**: QSA doesn't accept AWS documentation
**Solution**:

- Provide official AWS Artifact downloaded documents
- Share AWS compliance webpage references
- Engage AWS compliance team if needed

## **10. Additional AWS Compliance Resources**

### **10.1 AWS Compliance Center**

**Website**: <https://aws.amazon.com/compliance/>
**Resources**:

- Compliance program overviews
- Shared responsibility model guides
- Service-specific compliance information
- Best practice documentation

### **10.2 AWS Well-Architected Framework**

**Security Pillar**

- Security design principles
- AWS security best practices
- Architecture patterns for compliance
- Security assessment tools

### **10.3 AWS Support**

**AWS Support Plans**

- Business Support: Basic compliance guidance
- Enterprise Support: Dedicated compliance resources
- Enhanced Support: Compliance-specific assistance

**Engagement Options**

- Support cases for specific questions
- Architecture reviews for compliance
- Well-Architected Reviews including security
- Professional services for compliance projects

## **11. Integration with GYB Connect Compliance Program**

### **11.1 Policy Integration**

**Information Security Policy**

- Reference AWS shared responsibility model
- Include AWS compliance documentation requirements
- Define roles for AWS compliance management

**Risk Management Policy**

- Include AWS service risk assessments
- Plan for AWS compliance report updates
- Monitor AWS security notifications

### **11.2 Audit Preparation**

**Evidence Collection**

- AWS Artifact compliance reports
- Customer-implemented security controls
- Monitoring and logging configurations
- Incident response procedures and testing

**Documentation Package**

- Complete AWS compliance documentation
- GYB Connect security implementations
- Shared responsibility mapping
- Continuous monitoring evidence

## **12. Action Items and Next Steps**

### **12.1 Immediate Actions**

1. **Set up AWS Artifact access**
   - Configure IAM permissions
   - Accept customer agreements
   - Download current PCI DSS documentation

2. **Organize compliance documentation**
   - Create secure document storage
   - Implement access controls
   - Establish document lifecycle procedures

3. **Review shared responsibility**
   - Map AWS responsibilities to PCI DSS requirements
   - Identify customer implementation requirements
   - Plan control implementation and testing

### **12.2 Ongoing Activities**

1. **Regular document updates**
   - Monthly AWS Artifact checks
   - Quarterly document reviews
   - Annual compliance documentation refresh

2. **Audit preparation**
   - Maintain current evidence packages
   - Practice audit presentations
   - Keep QSA communication channels open

## **13. Compliance Evidence Checklist**

### **13.1 AWS Physical Security Evidence**

- [ ] AWS PCI DSS Attestation of Compliance (current year)
- [ ] AWS PCI DSS Shared Responsibility Matrix
- [ ] AWS SOC 2 Type II Report (latest available)
- [ ] AWS ISO 27001 Certificate
- [ ] AWS Data Center Security Overview whitepaper

### **13.2 Customer Implementation Evidence**

- [ ] Network security configuration documentation
- [ ] Access control implementation evidence
- [ ] Data encryption configuration proof
- [ ] Monitoring and logging implementation
- [ ] Incident response capability demonstration

## **14. Contact Information**

**AWS Support**

- AWS Support Console: <https://console.aws.amazon.com/support/>
- AWS Compliance Team: Contact through support case
- AWS Professional Services: For compliance assistance

**Internal GYB Connect Contacts**

- **Security Team**: <security@gybconnect.com>
- **Compliance Manager**: <compliance@gybconnect.com>
- **CTO**: [CTO Email]

---

**Document Owner**: Chief Technology Officer  
**Approved By**: [CTO Name]  
**Date**: [Approval Date]  
**Next Review**: January 2025

**For questions about this guide, contact**: <security@gybconnect.com>
