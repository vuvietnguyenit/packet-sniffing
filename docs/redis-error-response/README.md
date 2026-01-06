# sniff-redis-error-response

Sniffer Redis error is responed to client

## Usage

```sh
./packet-sniffer redis-error-response -h                                   
Record error reponses on Redis instance to clients base and export to Prometheus metrics base on libpcap

Usage:
  packet-sniffer redis-error-response [flags]

Flags:
  -e, --exporter-port int   Prometheus exporter port (default 2112)
  -h, --help                help for redis-error-response
  -i, --iface string        Network interface to monitor (default "eth0")
  -p, --port int            Redis port want to sniff (default 6379)
  -v, --verbose             Enable verbose logging
```

For example, if you want to sniff error responses from Redis server (on port 6379) at ens3 interface that are being sent back to clients.

```bash
# ./packet-sniffer redis-error-response -i ens3 -p 3306
{"time":"2025-11-17T11:33:05.547757911+07:00","level":"INFO","msg":"listening for MySQL error packets on","iface":"ens3","port":3306}
```

## Metrics

```sh
> curl localhost:2112/metrics
# HELP redis_error_response_count Number of Redis errors received
# TYPE redis_error_response_count counter
redis_error_response_count{error_code="ERR"} 7
redis_error_response_count{error_code="NOAUTH"} 16

```
