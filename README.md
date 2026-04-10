# CaddyStats

A lightweight traffic statistics module for [Caddy v2](https://caddyserver.com/) that tracks HTTP request metrics.

## Features

- **Request Counting**: Total number of HTTP requests processed
- **Upstream Bytes**: Total bytes received from clients (request body)
- **Downstream Bytes**: Total bytes sent to clients (response body)
- **Uptime Tracking**: Server uptime in seconds
- **JSON Stats Endpoint**: Exposes metrics via a simple REST API

## Installation

### Using xcaddy

```bash
xcaddy build --with github.com/yourusername/caddystat
```

### Using Docker

```dockerfile
FROM caddy:2-builder AS builder
RUN xcaddy build \
    --with github.com/yourusername/caddystat

FROM caddy:2
COPY --from=builder /usr/bin/caddy /usr/bin/caddy
```

Build and run:

```bash
docker build -t caddy-with-stats .
docker run -p 80:80 -p 443:443 caddy-with-stats
```

## Usage

Add the `stat` handler to your Caddyfile:

```caddyfile
example.com {
    handle {
        stat
        reverse_proxy localhost:3000
    }
}
```

Or in JSON config:

```json
{
  "apps": {
    "http": {
      "servers": {
        "srv0": {
          "listen": [":443"],
          "routes": [{
            "handle": [
              {
                "handler": "stat"
              },
              {
                "handler": "reverse_proxy",
                "upstreams": [{"dial": "localhost:3000"}]
              }
            ]
          }]
        }
      }
    }
  }
}
```

## Stats Endpoint

The module registers an admin endpoint at `/stats` that returns JSON metrics:

```bash
curl http://localhost:2019/stats
```

Response:

```json
{
  "upstream_bytes": 12345,
  "downstream_bytes": 67890,
  "request_count": 42,
  "uptime_seconds": 3600.5
}
```

## Module Info

- **Handler ID**: `http.handlers.stat`
- **Admin Endpoint**: `/stats`

## License

MIT
