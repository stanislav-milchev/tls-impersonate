# TLS Impersonator
![Coverage](https://img.shields.io/badge/Coverage-57.3%25-yellow)

# Introduction
A Man-in-the-Middle proxy that transforms requests in order to impersonate real browser request.
Developed for primary use in web scraping to bypass blocking.

# Features
- support for Chrome
- fast and resource friendly
- proxy support
- custom headers
- custom timeouts

- supply custom names for the 'dev' headers by passing env vars:
```
var => defaultValue
"TLS_PORT" => "8082"
"TLS_URL" => "x-tls-url"
"TLS_PROXY" => "x-tls-proxy"
"TLS_BUFFER" => "x-tls-buffer"
"TLS_REDIRECT" => "x-tls-allowredirect"
"TLS_TIMEOUT" => "x-tls-timeout"
```

# Coming soon
- Firefox impersonation
- more versions and headers in order to allow for ratation of browsers
