# Cloudflare Provider Examples

## Getting Started

These examples demonstrate how to use the Crossplane Cloudflare provider to manage Cloudflare resources.

## Installation

1. Install the provider package in your Crossplane cluster
2. Create a `ProviderConfig` with your Cloudflare API token (see `providerconfig/`)
3. Apply resource manifests (see resource-specific directories)

## Available Examples

- `providerconfig/` - ProviderConfig and Secret for Cloudflare API authentication
- `dns/` - DNS Record management

## Usage Pattern

1. Create a Kubernetes Secret with your Cloudflare API token
2. Create a ProviderConfig referencing the secret
3. Create managed resources (e.g., DNS Records, Zones) referencing the ProviderConfig
