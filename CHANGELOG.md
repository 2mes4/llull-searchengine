# Changelog

All notable changes to Llull Search Engine will be documented in this file.

## [Unreleased]

### Added
- **Multi-tenant index isolation**: Optional `LLULL_TENANT_PREFIX` env var or
  `X-Tenant` header to prefix all index names with a tenant identifier.
  Enables running multiple tenants on the same Llull instance without data
  collision (e.g. `makeyourcrew-space-abc123` instead of `space-abc123`).
