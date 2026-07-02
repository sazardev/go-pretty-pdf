# Security Policy

## Reporting a vulnerability

Please **do not** open a public issue for security vulnerabilities.

Instead, report them privately to the maintainers:

- Email: [security contact TBD]
- GitHub: Use the "Report a vulnerability" feature in the Security tab

We aim to respond within 5 business days and publish fixes as patch releases.

## Supported versions

| Version | Supported |
|---------|-----------|
| 0.1.x   | Yes       |

## Scope

`go-pretty-pdf` uses headless Chrome for PDF rendering. While we pass the
`--no-sandbox` flag for compatibility in CI environments, we recommend
running in sandboxed mode for production deployments that process
untrusted HTML content.

Note that custom CSS and HTML templates provided via configuration files
are injected directly into the rendered page. Only use trusted
CSS/template files.
