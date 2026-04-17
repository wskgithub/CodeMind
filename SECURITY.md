# Security Policy

## Supported Versions

The following versions of CodeMind are currently supported with security updates:

| Version | Supported          |
|---------|--------------------|
| 0.7.x   | :white_check_mark: Yes |
| < 0.7   | :x: No             |

---

## Reporting a Vulnerability

We take the security of CodeMind seriously. If you believe you have found a security vulnerability, please follow the guidelines below to report it responsibly.

### :warning: Please Do Not

- **Do not** open a public GitHub issue for a security vulnerability.
- **Do not** disclose the issue publicly until it has been addressed.

### :white_check_mark: How to Report

Please report security vulnerabilities via **[GitHub Private Vulnerability Reporting](https://github.com/wskgithub/CodeMind/security/advisories/new)**.

Alternatively, you may email the maintainers directly. Contact information can be found on the repository maintainer's GitHub profile.

### What to Include

To help us triage and resolve issues quickly, please include as much of the following information as possible:

- **Description**: A clear and concise description of the vulnerability.
- **Reproduction Steps**: Step-by-step instructions to reproduce the issue.
- **Impact**: An assessment of the potential impact or severity.
- **Affected Versions**: Which versions or components are affected.
- **Suggested Fix**: If you have ideas on how to fix the vulnerability, please share them.
- **Proof of Concept**: Any code, screenshots, or other materials that demonstrate the vulnerability.

### Response Timeline

Once we receive your report, you can expect the following:

| Stage | Timeline |
|-------|----------|
| **Acknowledgment** | Within 48 hours |
| **Initial Assessment** | Within 5 business days |
| **Fix & Validation** | As soon as practicable, depending on severity |
| **Public Disclosure** | Coordinated with the reporter after a fix is released |

We are committed to working with security researchers and the community to resolve vulnerabilities in a timely and transparent manner.

### Scope

The following areas are considered **in scope** for security reports:

- Authentication and authorization bypasses
- SQL injection or command injection
- Cross-site scripting (XSS)
- Cross-site request forgery (CSRF)
- Sensitive data exposure (API keys, passwords, tokens, JWT secrets)
- Server-side request forgery (SSRF)
- Privilege escalation
- Cryptographic weaknesses
- Insecure default configurations

The following are considered **out of scope**:

- Vulnerabilities in third-party dependencies (please report to the upstream project)
- Social engineering attacks
- Denial of service (DoS) attacks against infrastructure not under our control
- Issues related to third-party LLM providers
- Theoretical vulnerabilities without a practical proof of concept

### Security Best Practices for Deployment

When self-hosting CodeMind, we strongly recommend the following practices:

- Change default admin credentials immediately after the first login.
- Use strong, unique passwords for database and Redis connections.
- Ensure JWT secrets are at least 32 characters long and randomly generated.
- Enable HTTPS in production using a reverse proxy (e.g., Nginx, Traefik).
- Restrict database and Redis ports to internal networks only.
- Keep dependencies, Docker images, and the application up to date.

---

Thank you for helping keep CodeMind and its community safe!
