# NaiveStats

A lightweight traffic statistics module for [Caddy v2](https://caddyserver.com/), designed to count [NaiveProxy](https://github.com/klzgrad/naiveproxy) traffic.

## Features

- **Request Counting**: Total number of NaiveProxy CONNECT requests
- **Upstream Bytes**: Total bytes received from clients
- **Downstream Bytes**: Total bytes sent to clients
- **Uptime Tracking**: Server uptime in seconds
- **JSON Stats Endpoint**: Exposes metrics via Caddy Admin API

## Installation

### Using xcaddy

```bash
xcaddy build \
    --with github.com/caddyserver/forwardproxy@caddy2=github.com/klzgrad/forwardproxy@naive \
    --with github.com/haogong/CaddyStats
```

## Usage

Add the `naive_stat` handler to your Caddyfile, using a matcher to only count NaiveProxy (CONNECT) traffic:

```caddyfile
{
    admin :2019
}

example.com {
    tls user@example.com

    route {
        @naive method CONNECT
        naive_stat @naive
        forward_proxy {
            basic_auth user pass
            hide_ip
            hide_via
            probe_resistance
        }
        file_server {
            root /var/www/html
        }
    }
}
```

## Stats Endpoint

The module registers an admin endpoint at `/naive_stats`:

```bash
curl http://localhost:2019/naive_stats
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

- **Handler ID**: `http.handlers.naive_stat`
- **Caddyfile Directive**: `naive_stat`
- **Admin Endpoint**: `/naive_stats`

## License

MIT
