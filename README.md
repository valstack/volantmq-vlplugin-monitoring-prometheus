# Overview
[![pipeline status](https://gitlab.com/volantmq/vlplugin/monitoring/prometheus/badges/master/pipeline.svg)](https://gitlab.com/volantmq/vlplugin/monitoring/prometheus/commits/master)

## Config
Docker image of VolantMQ service comes with `prometheus` plugin
To enable plugin add `prometheus` value to list of enabled plugins. In config section listening port and path can be specified
```yaml
plugins:
  enabled:
    - prometheus
  config:
    monitoring:
      - backend: prometheus
        config:
          path: "/metrics"
          port: 8080
```

## Grafana dashboard
Check out examples
