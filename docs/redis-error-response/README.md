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

The result

```text
{"time":"2026-01-06T13:36:48.767124501+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":46948,"length":55,"error_message":"-ERR Protocol error: unauthenticated multibulk length\r\n"}
{"time":"2026-01-06T13:36:49.759150356+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":46956,"length":34,"error_message":"-NOAUTH Authentication required.\r\n"}
{"time":"2026-01-06T13:36:50.771151317+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":46964,"length":34,"error_message":"-NOAUTH Authentication required.\r\n"}
{"time":"2026-01-06T13:36:51.763089382+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":46972,"length":55,"error_message":"-ERR Protocol error: unauthenticated multibulk length\r\n"}
{"time":"2026-01-06T13:36:52.771195679+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":46974,"length":76,"error_message":"-ERR unknown command 'THIS', with args beginning with: 'IS' 'NOT' 'REDIS' \r\n"}
{"time":"2026-01-06T13:36:53.771122482+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":46988,"length":34,"error_message":"-NOAUTH Authentication required.\r\n"}
{"time":"2026-01-06T13:36:54.771171408+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":47004,"length":55,"error_message":"-ERR Protocol error: unauthenticated multibulk length\r\n"}
{"time":"2026-01-06T13:36:55.783103908+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":47010,"length":34,"error_message":"-NOAUTH Authentication required.\r\n"}
{"time":"2026-01-06T13:36:57.783163952+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":47024,"length":55,"error_message":"-ERR Protocol error: unauthenticated multibulk length\r\n"}
{"time":"2026-01-06T13:36:58.77511459+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":50254,"length":34,"error_message":"-NOAUTH Authentication required.\r\n"}
{"time":"2026-01-06T13:36:59.78308653+07:00","level":"INFO","msg":"MySQL ERR packet","src_ip":"10.194.60.166","src_port":6379,"dst_ip":"10.194.60.90","dst_port":50266,"length":55,"error_message":"-ERR Protocol error: unauthenticated multibulk length\r\n"}

```
## Metrics

```sh
> curl localhost:2112/metrics
# HELP redis_error_response_count Number of Redis errors received
# TYPE redis_error_response_count counter
redis_error_response_count{error_code="ERR"} 7
redis_error_response_count{error_code="NOAUTH"} 16

```
