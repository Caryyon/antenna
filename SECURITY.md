# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 1.x     | :white_check_mark: |

## Reporting a Vulnerability

If you find a security issue, please **do not** open a public issue.

Instead, email **cary@caryyon.com** with:
- Description of the vulnerability
- Steps to reproduce
- Potential impact

You'll get a response within 48 hours. We'll work with you to understand and fix the issue before any public disclosure.

## Security Considerations

Antenna reads session data from `~/.openclaw/`. It does not:
- Send data to external servers
- Require network access (except for the embedded WebKit view)
- Store any credentials

The app runs entirely locally on your machine.
