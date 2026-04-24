# Security Policy

## Supported Versions

Only the latest release is actively supported with security fixes.

| Version | Supported |
| ------- | --------- |
| 3.2.x   | Yes       |
| < 3.2   | No        |

## Reporting a Vulnerability

**Do not open a public issue for security vulnerabilities.**

Please report security vulnerabilities via
[GitHub Security Advisories](https://github.com/tfutils/tfenv/security/advisories/new).

You can expect:

- **Acknowledgement** within 48 hours
- **Status update** within 7 days
- **Fix or mitigation plan** within 30 days for confirmed vulnerabilities

If you do not receive a response within 48 hours, please email the maintainer
directly at mike.peachey@bjss.com.

## Scope

tfenv downloads and installs Terraform binaries from HashiCorp's release
servers. Security concerns relevant to this project include:

- Signature and checksum verification bypass
- Path traversal or injection via version strings
- Arbitrary code execution through crafted `.terraform-version` files
- Supply chain attacks via the tfenv installation or update mechanism
