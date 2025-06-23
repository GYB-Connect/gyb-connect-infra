# **GYB Connect Risk Management Policy**

**Document Version**: 1.0  
**Effective Date**: January 2024  
**Review Date**: January 2025  
**Classification**: Internal Use Only

## **1. Purpose and Scope**

This Risk Management Policy establishes a systematic approach for identifying, assessing, and managing information security risks that could impact the confidentiality, integrity, or availability of cardholder data and supporting systems.

**Scope**: All business processes, systems, and technologies that handle, store, process, or transmit cardholder data or connect to the Cardholder Data Environment (CDE).

## **2. Risk Management Framework**

### **2.1 Risk Management Process**

Our risk management process follows a continuous cycle:

1. **Risk Identification**: Systematically identify potential security threats and vulnerabilities
2. **Risk Assessment**: Analyze the likelihood and impact of identified risks
3. **Risk Treatment**: Implement appropriate controls to mitigate, transfer, accept, or avoid risks
4. **Risk Monitoring**: Continuously monitor and review risks and control effectiveness
5. **Risk Communication**: Report risks to stakeholders and maintain risk awareness

### **2.2 Risk Governance**

**Risk Owner**: Chief Technology Officer (CTO)

- Ultimate accountability for risk management program
- Approval of risk treatment decisions
- Resource allocation for risk mitigation

**Risk Manager**: Security Team Lead

- Day-to-day risk management activities
- Risk assessments and analysis
- Risk register maintenance

**Risk Committee**: Executive team + Security team

- Quarterly risk review meetings
- Risk treatment decision approval
- Risk appetite and tolerance setting

## **3. Risk Assessment Methodology**

### **3.1 Risk Identification**

**Threat Sources**

- External threats (cybercriminals, nation-states, hacktivists)
- Internal threats (malicious insiders, negligent employees)
- Environmental threats (natural disasters, infrastructure failures)
- Technical threats (system failures, software vulnerabilities)

**Vulnerability Categories**

- Network vulnerabilities (firewall misconfigurations, open ports)
- System vulnerabilities (unpatched software, weak configurations)
- Application vulnerabilities (code flaws, injection attacks)
- Physical vulnerabilities (unauthorized access, equipment theft)
- Process vulnerabilities (inadequate procedures, human error)

**Asset Classification**

- **Critical**: Cardholder data, authentication systems, payment processing
- **High**: Customer data, financial records, administrative systems
- **Medium**: Business applications, employee data, internal communications
- **Low**: Public information, development systems, documentation

### **3.2 Risk Analysis**

**Likelihood Assessment**

- **Very High (5)**: >80% probability within 12 months
- **High (4)**: 60-80% probability within 12 months
- **Medium (3)**: 40-60% probability within 12 months
- **Low (2)**: 20-40% probability within 12 months
- **Very Low (1)**: <20% probability within 12 months

**Impact Assessment**

- **Very High (5)**: Catastrophic business impact, regulatory action, >$1M loss
- **High (4)**: Major business disruption, significant legal issues, $100K-$1M loss
- **Medium (3)**: Moderate business impact, compliance violations, $10K-$100K loss
- **Low (2)**: Minor business impact, internal issues, $1K-$10K loss
- **Very Low (1)**: Negligible impact, minimal business disruption, <$1K loss

**Risk Calculation**
Risk Level = Likelihood × Impact

| Risk Score | Risk Level | Required Action |
|------------|------------|----------------|
| 20-25 | Critical | Immediate action required |
| 15-19 | High | Action required within 30 days |
| 10-14 | Medium | Action required within 90 days |
| 5-9 | Low | Monitor and review quarterly |
| 1-4 | Very Low | Accept risk or monitor annually |

### **3.3 Risk Assessment Process**

**Annual Comprehensive Assessment**

- Complete review of all systems and processes
- Threat landscape analysis
- Vulnerability assessments
- Business impact analysis
- Risk register update

**Quarterly Risk Reviews**

- Review existing risks and controls
- Identify new risks
- Assess control effectiveness
- Update risk scores
- Report to executive team

**Ad-hoc Assessments**

- Major system changes
- New technology implementations
- Security incidents
- Regulatory changes
- Third-party integrations

## **4. Risk Treatment Strategies**

### **4.1 Risk Treatment Options**

**Mitigate**

- Implement controls to reduce likelihood or impact
- Most common approach for operational risks
- Examples: Firewalls, encryption, access controls

**Transfer**

- Share risk with third parties
- Insurance, outsourcing, contracts
- Examples: Cyber insurance, cloud provider shared responsibility

**Accept**

- Acknowledge risk and continue operations
- Appropriate for low-level risks
- Requires formal acceptance and documentation

**Avoid**

- Eliminate the risk by changing business processes
- Remove vulnerable systems or processes
- Examples: Discontinue risky services, change technology

### **4.2 Control Categories**

**Preventive Controls**

- Prevent security incidents from occurring
- Examples: Firewalls, access controls, encryption
- Priority for high-likelihood risks

**Detective Controls**

- Identify security incidents when they occur
- Examples: Monitoring, logging, intrusion detection
- Essential for sophisticated threats

**Corrective Controls**

- Respond to and recover from security incidents
- Examples: Backup systems, incident response, business continuity
- Critical for business resilience

**Compensating Controls**

- Alternative controls when standard controls aren't feasible
- Must provide equivalent protection
- Require PCI DSS QSA approval

## **5. Risk Appetite and Tolerance**

### **5.1 Risk Appetite Statement**

GYB Connect maintains a **conservative risk appetite** for information security risks that could impact cardholder data security or business operations. We prioritize security over convenience and will invest in appropriate controls to maintain this posture.

### **5.2 Risk Tolerance Levels**

**Zero Tolerance**

- Unauthorized access to cardholder data
- Non-compliance with PCI DSS requirements
- Data breaches or security incidents

**Low Tolerance**

- System availability issues
- Performance degradation
- Minor compliance violations

**Moderate Tolerance**

- Development environment issues
- Non-critical system outages
- Internal process inefficiencies

### **5.3 Risk Escalation Thresholds**

| Risk Level | Notification | Approval Required |
|------------|-------------|------------------|
| Critical | CTO immediate | Board approval |
| High | CTO within 24 hours | Executive team |
| Medium | Weekly report | Risk Manager |
| Low | Monthly report | Department head |

## **6. Risk Monitoring and Reporting**

### **6.1 Risk Metrics**

**Leading Indicators**

- Number of vulnerabilities identified
- Security control coverage
- Training completion rates
- Vendor security assessments completed

**Lagging Indicators**

- Number of security incidents
- Compliance violations
- Business disruption hours
- Financial losses from security events

### **6.2 Risk Reporting Schedule**

**Real-time**

- Critical risk alerts
- Security incident notifications
- Control failure alerts

**Weekly**

- Risk dashboard updates
- New risk identification
- Control effectiveness metrics

**Monthly**

- Risk register updates
- Executive risk summary
- Trend analysis

**Quarterly**

- Comprehensive risk assessment
- Risk committee meetings
- Board reporting

**Annually**

- Complete risk program review
- Risk appetite reassessment
- Regulatory compliance review

### **6.3 Key Risk Indicators (KRIs)**

**Technical KRIs**

- Critical vulnerabilities >30 days old
- Failed backup attempts
- Unauthorized access attempts
- System availability <99.9%

**Process KRIs**

- Overdue access reviews
- Unresolved audit findings
- Training completion <95%
- Vendor assessments overdue

**Business KRIs**

- Revenue impact from security issues
- Customer complaints about security
- Regulatory inquiries
- Insurance premium increases

## **7. Third-Party Risk Management**

### **7.1 Vendor Risk Assessment**

**Pre-engagement Assessment**

- Security questionnaire completion
- Financial stability review
- Compliance certification verification
- Reference checks

**Due Diligence Requirements**

- SOC 2 Type II reports
- PCI DSS compliance (for payment processors)
- ISO 27001 certification (preferred)
- Penetration test results

**Risk Classification**

- **Critical**: Handle cardholder data or critical business functions
- **High**: Access to sensitive data or systems
- **Medium**: Limited access to business data
- **Low**: No access to sensitive data or systems

### **7.2 Ongoing Vendor Monitoring**

**Quarterly Reviews**

- Security posture assessment
- Compliance status verification
- Incident notification review
- Performance metrics analysis

**Annual Assessments**

- Comprehensive security review
- Contract renewal evaluation
- Risk classification update
- Alternative vendor evaluation

### **7.3 Vendor Incident Response**

**Notification Requirements**

- Security incidents within 24 hours
- Compliance violations immediately
- Service disruptions within 2 hours
- Data breaches immediately

**Response Procedures**

- Activate vendor incident response plan
- Assess impact on GYB Connect
- Implement additional controls if needed
- Document lessons learned

## **8. Risk Management Technology**

### **8.1 AWS Security Services**

**Security Hub**

- Centralized security findings
- Compliance status monitoring
- Risk prioritization
- Automated remediation

**GuardDuty**

- Threat detection and analysis
- Malicious activity identification
- Risk scoring and prioritization
- Automated response capabilities

**Inspector**

- Vulnerability assessment
- Software composition analysis
- Risk-based patching prioritization
- Continuous monitoring

**Config**

- Configuration compliance monitoring
- Drift detection
- Automated remediation
- Compliance reporting

### **8.2 Risk Management Tools**

**GRC Platform Features**

- Risk register management
- Assessment workflow automation
- Control testing tracking
- Executive reporting dashboards

**Integration Capabilities**

- AWS security service integration
- Vulnerability scanner feeds
- Threat intelligence integration
- SIEM/SOAR platform connectivity

## **9. Training and Awareness**

### **9.1 Risk Management Training**

**All Personnel**

- Annual risk awareness training
- Role-specific risk responsibilities
- Incident reporting procedures
- Risk escalation processes

**Risk Management Team**

- Quarterly risk assessment training
- Risk analysis methodology
- Control evaluation techniques
- Regulatory requirement updates

**Executive Team**

- Risk governance training
- Strategic risk decision making
- Risk appetite setting
- Board reporting requirements

### **9.2 Risk Communication**

**Risk Culture**

- Open communication about risks
- No-blame incident reporting
- Continuous improvement mindset
- Risk-informed decision making

**Awareness Programs**

- Monthly risk newsletters
- Quarterly risk lunch-and-learns
- Annual risk management day
- Incident case study reviews

## **10. Compliance and Audit**

### **10.1 Regulatory Compliance**

**PCI DSS Requirements**

- Annual risk assessments (Req. 12.2)
- Risk-based security controls
- Compensating control validation
- Compliance evidence collection

**Other Frameworks**

- NIST Cybersecurity Framework
- ISO 27001/27002
- SOC 2 controls
- Industry best practices

### **10.2 Internal Audit**

**Annual Audit Program**

- Risk management process effectiveness
- Control implementation verification
- Risk register accuracy
- Compliance with policy requirements

**Audit Evidence**

- Risk assessment documentation
- Control testing results
- Remediation tracking
- Executive reporting records

## **11. Policy Maintenance**

### **11.1 Review and Updates**

**Annual Policy Review**

- Risk management framework effectiveness
- Threat landscape changes
- Regulatory requirement updates
- Technology changes impact

**Quarterly Process Review**

- Risk assessment methodology
- Risk treatment effectiveness
- Reporting accuracy
- Stakeholder feedback

### **11.2 Continuous Improvement**

**Improvement Opportunities**

- Process automation
- Tool integration
- Reporting enhancement
- Training effectiveness

**Best Practice Adoption**

- Industry standard updates
- Peer organization practices
- Regulatory guidance
- Technology innovations

## **12. Related Documents**

- [Information Security Policy](./Information_Security_Policy.md)
- [Incident Response Plan](./Incident_Response_Plan.md)
- [Access Control Policy](./Access_Control_Policy.md)
- [Business Continuity Plan](./Business_Continuity_Plan.md)

---

**Document Owner**: Chief Technology Officer  
**Approved By**: [CTO Name]  
**Date**: [Approval Date]  
**Next Review**: January 2025

**For questions about this policy, contact**: <security@gybconnect.com>
