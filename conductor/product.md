# Product Guide: Apicurio Registry Terraform Provider

## Overview
This project is a Terraform Provider for managing Apicurio Registry instances. It allows infrastructure as code (IaC) management of schemas, API designs, and compatibility rules.

## Target Users
- **DevOps Engineers:** Infrastructure engineers automating Apicurio Registry deployments and configuration.

## Key Resources
The provider will enable management of the following key Apicurio Registry resources:
- **Artifacts:** Resources for managing schema artifacts (Avro, Protobuf, JSON).
- **Groups & Metadata:** Resources for managing artifact groups and associated metadata.
- **Rules:** Resources for configuring global or artifact-specific compatibility rules.

## Primary Use Case
- **Multi-Environment Management:** Managing complex, multi-environment schema registries to ensure consistency across development, staging, and production.