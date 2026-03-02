# Crossplane Provider Cloudflare

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![GitHub release](https://img.shields.io/github/release/millstonehq/provider-upjet-cloudflare.svg)](https://github.com/millstonehq/provider-upjet-cloudflare/releases)

A Crossplane provider for managing Cloudflare infrastructure declaratively using Kubernetes-style APIs.

## Overview

This provider enables you to manage Cloudflare resources through Crossplane, bringing GitOps-style infrastructure management to your Cloudflare zones and DNS records.

Built with **Upjet v2** wrapping **Terraform cloudflare/cloudflare v5.17.0**.

### Supported Resources

- **DNS Record** (`cloudflare_dns_record`) - Manage DNS records (A, AAAA, CNAME, MX, TXT, etc.)
- **Zone** (`cloudflare_zone`) - Manage Cloudflare zones

## Installation

### Prerequisites

- Kubernetes cluster with Crossplane installed (v1.14.0+)
- Cloudflare account with API access
- Crossplane 2.0+ installed in your cluster

### Install the Provider

```bash
# Create the provider
kubectl apply -f - <<EOF
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-cloudflare
spec:
  package: ghcr.io/millstonehq/provider-cloudflare:latest
EOF

# Verify installation
kubectl get providers
```

### Configure Authentication

1. **Create a Cloudflare API Token**

   Visit the [Cloudflare Dashboard](https://dash.cloudflare.com/profile/api-tokens) and create a new API token with appropriate permissions (Zone:DNS:Edit, Zone:Zone:Read).

2. **Create a Kubernetes Secret**

   ```bash
   kubectl create secret generic cloudflare-creds \
     --namespace crossplane-system \
     --from-literal=api_token='your-cloudflare-api-token'
   ```

3. **Create a ProviderConfig**

   ```bash
   kubectl apply -f - <<EOF
   apiVersion: cloudflare.upbound.io/v1beta1
   kind: ProviderConfig
   metadata:
     name: default
   spec:
     credentials:
       source: Secret
       secretRef:
         name: cloudflare-creds
         namespace: crossplane-system
         key: api_token
   EOF
   ```

## Usage Examples

### DNS Record Management

```yaml
apiVersion: dns.cloudflare.upbound.io/v1alpha1
kind: Record
metadata:
  name: example-a-record
spec:
  forProvider:
    zoneId: "your-zone-id"
    name: "app.example.com"
    type: "A"
    content: "192.0.2.1"
    ttl: 3600
    proxied: true
  providerConfigRef:
    name: default
```

### CNAME Record

```yaml
apiVersion: dns.cloudflare.upbound.io/v1alpha1
kind: Record
metadata:
  name: example-cname
spec:
  forProvider:
    zoneId: "your-zone-id"
    name: "www.example.com"
    type: "CNAME"
    content: "app.example.com"
    proxied: true
  providerConfigRef:
    name: default
```

### MX Record

```yaml
apiVersion: dns.cloudflare.upbound.io/v1alpha1
kind: Record
metadata:
  name: example-mx
spec:
  forProvider:
    zoneId: "your-zone-id"
    name: "example.com"
    type: "MX"
    content: "mail.example.com"
    priority: 10
  providerConfigRef:
    name: default
```

## Development

### Building from Source

This provider uses [Earthly](https://earthly.dev) for building and testing.

```bash
# Generate code
earthly +generate

# Build the provider
earthly +build

# Run tests
earthly +test

# Test with examples
earthly +test-examples

# Run all tests (unit + examples)
earthly +test-all

# Build and push images (requires authentication)
earthly --push +push
```

### Local Development

```bash
# Build provider package locally
earthly +package-local

# Install in your cluster
kubectl apply -f examples/providerconfig/
```

## Architecture

This provider is built using:
- **Upjet v2.0.0** - Code generation framework for Terraform-based Crossplane providers
- **Crossplane Runtime v2.0.0** - Core Crossplane functionality
- **Terraform Provider Cloudflare v5.17.0** - Underlying Terraform provider

### Authentication Methods

The provider supports two authentication methods:

1. **API Token** (Recommended)
   ```yaml
   stringData:
     api_token: "your-cloudflare-api-token"
   ```

2. **API Key + Email** (Legacy)
   ```yaml
   stringData:
     api_key: "your-cloudflare-api-key"
     email: "your-cloudflare-email"
   ```

## Community & Contributing

We welcome contributions from the community! Whether you're fixing bugs, adding features, or improving documentation, your help is appreciated.

### How to Contribute

1. **Fork the repository** on GitHub
2. **Create a feature branch** (`git checkout -b feature/amazing-feature`)
3. **Make your changes** and commit them (`git commit -m 'feat: add amazing feature'`)
4. **Push to your branch** (`git push origin feature/amazing-feature`)
5. **Open a Pull Request**

Please read our [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines on:
- Development setup and building from source
- Code style and conventions
- Testing requirements
- PR submission process

### Getting Help

- [Documentation](https://github.com/millstonehq/provider-upjet-cloudflare/tree/main/examples)
- [Report a Bug](https://github.com/millstonehq/provider-upjet-cloudflare/issues/new?labels=bug)
- [Request a Feature](https://github.com/millstonehq/provider-upjet-cloudflare/issues/new?labels=enhancement)
- [Discussions](https://github.com/millstonehq/provider-upjet-cloudflare/discussions)

### Code of Conduct

This project follows the [Contributor Covenant Code of Conduct](https://www.contributor-covenant.org/version/2/1/code_of_conduct/). By participating, you are expected to uphold this code.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

Copyright 2025 Millstone Partners, LLC

## Support

For issues and questions:
- GitHub Issues: https://github.com/millstonehq/provider-upjet-cloudflare/issues
- Documentation: https://github.com/millstonehq/provider-upjet-cloudflare/tree/main/examples
- Discussions: https://github.com/millstonehq/provider-upjet-cloudflare/discussions

## References

- [Cloudflare API Documentation](https://developers.cloudflare.com/api)
- [Cloudflare Terraform Provider](https://registry.terraform.io/providers/cloudflare/cloudflare/latest/docs)
- [Crossplane Documentation](https://docs.crossplane.io)
- [Upjet Documentation](https://github.com/crossplane/upjet)
