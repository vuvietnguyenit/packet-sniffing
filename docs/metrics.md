# Metric exporter

The labels of metric will be added by environment variables have the prefix MYSQL_ERROR_ECHO_METRIC_*. For example:

```text
MYSQL_ERROR_ECHO_METRIC_CLUSTER_NAME
MYSQL_ERROR_ECHO_METRIC_K8S_NAMESPACE
MYSQL_ERROR_ECHO_METRIC_BLAH_BLAH
```

The suffix after MYSQL_ERROR_ECHO_METRIC_ will be collected as label key of metrics.

It is used to identify MySQL instance that is exposed metrics, and then we can use it to do label grouping in the future.

## Usage

```bash
> MYSQL_ERROR_ECHO_METRIC_CLUSTER=mysql-prod MYSQL_ERROR_ECHO_METRIC_K8S_NAMESPACE=default ./mysql-error-echo --iface ens3 -v

{"time":"2025-12-02T11:13:19.84991977+07:00","level":"INFO","msg":"metric info","labels_included":["error_code","CLUSTER","K8S_NAMESPACE","NAMESPACE"]}
{"time":"2025-12-02T11:13:19.850061864+07:00","level":"INFO","msg":"starting metrics server","port":2112}
{"time":"2025-12-02T11:13:19.880302055+07:00","level":"INFO","msg":"listening for MySQL error packets on","iface":"ens3","port":3306}

```

## Metrics

```sh
> curl vunv-proj-sb.dev.virt:2112/metrics
# HELP mysql_error_response_count Number of MySQL errors received
# TYPE mysql_error_response_count counter
mysql_error_response_count{CLUSTER="mysql-prod",K8S_NAMESPACE="default",error_code="08S01"} 2
```
