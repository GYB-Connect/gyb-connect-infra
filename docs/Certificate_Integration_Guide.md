# Certificate Integration Guide

## Overview

This guide explains how your ACM certificate is integrated into the GYB Connect infrastructure to provide PCI DSS compliant TLS 1.2+ encryption for API endpoints.

## Certificate Configuration

### Your Certificate Details

- **ARN**: `arn:aws:acm:us-west-1:302004535198:certificate/179baacb-3105-4808-91f5-01090c89eb71`
- **Region**: `us-west-1`
- **Account**: `302004535198`

### Environment Variable Setup

The certificate ARN is configured via the `ACM_CERTIFICATE_ARN` environment variable, which is automatically set in:

- `deploy.sh` script
- `setup-env.sh` script
- `.env` file (created by setup script)

## Domain Mapping

The infrastructure automatically selects domain names based on the deployment environment:

| Environment | Domain Name |
|------------|-------------|
| Development (`dev`) | `api-dev.gybconnect.com` |
| Production (`prod`) | `api.gybconnect.com` |

### Customizing Domain Names

To use different domain names, update the domain configuration in `gyb_connect.go`:

```go
// Set domain name based on environment
if environment == stacks.PROD_ENV {
    domainName = "api.yourdomain.com" // Update this
} else {
    domainName = "api-dev.yourdomain.com" // Update this
}
```

## How Certificate Integration Works

### 1. Certificate Import

```go
// Get certificate ARN from environment variable
certArn := os.Getenv("ACM_CERTIFICATE_ARN")
if certArn != "" {
    certificate = awscertificatemanager.Certificate_FromCertificateArn(
        app,
        jsii.String("ApiCertificate"),
        jsii.String(certArn),
    )
}
```

### 2. Custom Domain Creation

```go
customDomain = awsapigatewayv2.NewDomainName(stack, jsii.String("GybConnectCustomDomain"), &awsapigatewayv2.DomainNameProps{
    DomainName:  jsii.String(props.DomainName),
    Certificate: props.Certificate,
    // PCI DSS Req 4.1: Enforce TLS 1.2+ - reject connections using older protocols
    SecurityPolicy: awsapigatewayv2.SecurityPolicy_TLS_1_2,
})
```

### 3. API Mapping

```go
awsapigatewayv2.NewApiMapping(stack, jsii.String("GybConnectApiMapping"), &awsapigatewayv2.ApiMappingProps{
    Api:        api,
    DomainName: customDomain,
    Stage:      stage,
})
```

## Deployment Process

### 1. Environment Setup

```bash
# Run the setup script to configure environment variables
./setup-env.sh

# Or manually set the environment variable
export ACM_CERTIFICATE_ARN="arn:aws:acm:us-west-1:302004535198:certificate/179baacb-3105-4808-91f5-01090c89eb71"
```

### 2. Deploy Infrastructure

```bash
# Deploy development environment
./deploy.sh dev

# Deploy production environment
./deploy.sh prod
```

### 3. Verify Deployment

After deployment, the script will output:

```
Custom Domain:
  Domain: api-dev.gybconnect.com (or api.gybconnect.com for prod)
  Alias Target: d-xxxxxxxxxx.execute-api.us-west-1.amazonaws.com
```

## DNS Configuration

### Required DNS Records

After deployment, you need to create DNS records to point your domains to the API Gateway:

```bash
# For development
api-dev.gybconnect.com -> CNAME -> d-xxxxxxxxxx.execute-api.us-west-1.amazonaws.com

# For production  
api.gybconnect.com -> CNAME -> d-xxxxxxxxxx.execute-api.us-west-1.amazonaws.com
```

### DNS Setup Steps

1. Get the alias target from deployment output
2. Log into your DNS provider (Route 53, Cloudflare, etc.)
3. Create a CNAME record pointing your domain to the alias target
4. Wait for DNS propagation (usually 5-15 minutes)

## Security Features

### TLS 1.2+ Enforcement

- **Security Policy**: `TLS_1_2` - Rejects connections using TLS 1.0 or 1.1
- **PCI DSS Compliance**: Meets Requirement 4.1 for strong cryptography
- **Cipher Suites**: Uses only secure cipher suites approved by AWS

### Certificate Validation

- **Domain Validation**: Certificate must be valid for the configured domains
- **Region Requirement**: Certificate must be in the same region as API Gateway
- **Automatic Renewal**: ACM handles certificate renewal automatically

## Testing and Verification

### 1. Test TLS Version

```bash
# Test that TLS 1.1 is rejected
openssl s_client -connect api-dev.gybconnect.com:443 -tls1_1

# Test that TLS 1.2 is accepted
openssl s_client -connect api-dev.gybconnect.com:443 -tls1_2
```

### 2. Check Certificate

```bash
# Verify certificate details
openssl s_client -connect api-dev.gybconnect.com:443 -showcerts
```

### 3. Test API Endpoints

```bash
# Test HTTPS endpoint
curl -v https://api-dev.gybconnect.com/

# Verify security headers
curl -I https://api-dev.gybconnect.com/
```

## Troubleshooting

### Common Issues

1. **Certificate Not Found Error**:
   - Verify the certificate ARN is correct
   - Ensure the certificate is in the us-west-1 region
   - Check that the certificate status is "Issued"

2. **Domain Validation Failed**:
   - Verify the certificate covers the domain you're using
   - Check that the domain matches exactly (including subdomains)
   - Ensure wildcard certificates include the specific subdomain

3. **DNS Resolution Issues**:
   - Verify CNAME record is correctly configured
   - Check DNS propagation using `dig` or `nslookup`
   - Wait for full DNS propagation (up to 48 hours)

4. **TLS Handshake Failures**:
   - Verify the security policy is TLS 1.2
   - Check that client supports TLS 1.2+
   - Ensure certificate chain is complete

### Debug Commands

```bash
# Check certificate status
aws acm describe-certificate --certificate-arn "arn:aws:acm:us-west-1:302004535198:certificate/179baacb-3105-4808-91f5-01090c89eb71"

# Check API Gateway domain
aws apigatewayv2 get-domain-names --region us-west-1

# Test DNS resolution
dig api-dev.gybconnect.com
nslookup api-dev.gybconnect.com
```

## Monitoring and Maintenance

### CloudWatch Metrics

Monitor these API Gateway metrics:

- `4XXError` - Client errors
- `5XXError` - Server errors  
- `Latency` - Response times
- `Count` - Request volume

### Certificate Monitoring

- ACM automatically renews certificates before expiration
- Set up CloudWatch alarms for certificate expiration warnings
- Monitor certificate validation status

### Security Monitoring

- Monitor for TLS version downgrade attempts
- Track unusual traffic patterns
- Set up alerts for certificate-related errors

## Next Steps

1. **Deploy the infrastructure** using `./deploy.sh`
2. **Configure DNS records** with your DNS provider
3. **Test the endpoints** to verify TLS 1.2+ enforcement
4. **Set up monitoring** for certificate and API health
5. **Document your specific domain configuration** for your team

## PCI DSS Compliance

This certificate integration provides:

- **Requirement 4.1**: Strong cryptography for data transmission
- **Requirement 2.2.7**: Encrypted administrative access protocols
- **Requirement 6.4.2**: Secure communication protocols for web applications

The infrastructure now meets PCI DSS requirements for protecting cardholder data in transit with industry-standard encryption.
