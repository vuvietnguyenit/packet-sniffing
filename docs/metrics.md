# Metric exporter

This tool helps generate Prometheus monitoring metrics.

Before exposing the metrics that are scraped by Prometheus server. We had provided two environment variables. They are used to identify MySQL cluster that is exposed metrics

Environment variables:

```text
- CLUSTER_NAME 
- NAMESPACE # will be good to provide it if your MySQL cluster running in Kubernetes namespace
```

## Usage

```bash
NAMESPACE=default CLUSTER_NAME=my-cluster ./mysql-error-echo --iface ens3 -v --exporter-port 9119

{"time":"2025-12-01T16:34:28.314466776+07:00","level":"INFO","msg":"starting metrics server","port":9119}
{"time":"2025-12-01T16:34:28.348339361+07:00","level":"INFO","msg":"listening for MySQL error packets on","iface":"ens3","port":3306}
{"time":"2025-12-01T16:34:41.14820936+07:00","level":"DEBUG","msg":"scrape request","from":"10.196.6.35:52208","path":"/metrics"}

```

## Metrics

```text
# curl localhost:9119/metrics
# HELP mysql_error_response_count Number of MySQL error responses
# TYPE mysql_error_response_count counter
# The cluster_name,namespace labels are mapped with env.
mysql_error_response_count{cluster_name="my-cluster",error_code="08S01",namespace="default"} 1
```
