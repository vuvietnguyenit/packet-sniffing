# Metric exporter

The labels of metric will be added by environment variables have the prefix MYSQL_ERROR_ECHO_METRIC_*. For example:

```text
MYSQL_ERROR_ECHO_METRIC_CLUSTER_NAME
MYSQL_ERROR_ECHO_METRIC_K8S_NAMESPACE
MYSQL_ERROR_ECHO_METRIC_BLAH_BLAH
```

The suffix after MYSQL_ERROR_ECHO_METRIC_ will be collected as label key of metrics.

It is used to identify MySQL instance that is exposed metrics, and then we can use it to do label grouping in the future. See [Example sidecar container](./examples/deployment.yaml#L87) if you're deploying it as sidecar container.

## Usage

```bash
> MYSQL_ERROR_ECHO_METRIC_CLUSTER=mysql-prod MYSQL_ERROR_ECHO_METRIC_K8S_NAMESPACE=default ./mysql-error-echo --iface ens3 -v

{"time":"2025-12-05T12:33:25.923917421+07:00","level":"INFO","msg":"metric info","labels_included":["error_code","state_code","CLUSTER","K8S_NAMESPACE","K8S_ENV"]}
{"time":"2025-12-05T12:33:25.924155461+07:00","level":"INFO","msg":"starting metrics server","port":2112}
{"time":"2025-12-05T12:33:25.945058926+07:00","level":"INFO","msg":"listening for MySQL error packets on","iface":"ens3","port":3306}

```

## Metrics

```sh
> curl localhost:2112/metrics
# HELP mysql_error_response_count Number of MySQL errors received
# TYPE mysql_error_response_count counter
mysql_error_response_count{CLUSTER="vunv-mysql-cluster",K8S_ENV="dev",K8S_NAMESPACE="default",error_code="1158",state_code="08S01"} 2
```
