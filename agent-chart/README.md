# mission-control-agent

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.10](https://img.shields.io/badge/AppVersion-0.0.10-informational?style=flat-square)

A Helm chart for flanksource mission control agent

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Flanksource |  |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://flanksource.github.io/charts | canary-checker | 1.1.2-beta.115 |
| https://flanksource.github.io/charts | config-db | 0.0.989 |
| https://flanksource.github.io/charts | pushTelemetry(mission-control-watchtower) | 0.1.28 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| canary-checker.db.external.create | bool | `false` |  |
| canary-checker.db.external.enabled | bool | `true` |  |
| canary-checker.db.external.secretKeyRef.key | string | `"DB_URL"` |  |
| canary-checker.db.external.secretKeyRef.name | string | `"incident-commander-postgres"` |  |
| canary-checker.db.runMigrations | bool | `false` |  |
| canary-checker.flanksource-ui.enabled | bool | `false` |  |
| canary-checker.image.type | string | `"full"` |  |
| canary-checker.logLevel | string | `"{{.Values.global.logLevel}}"` |  |
| config-db.db.embedded.persist | bool | `false` |  |
| config-db.db.external.enabled | bool | `true` |  |
| config-db.db.external.secretKeyRef.key | string | `"DB_URL"` |  |
| config-db.db.external.secretKeyRef.name | string | `"incident-commander-postgres"` |  |
| config-db.db.runMigrations | bool | `false` |  |
| config-db.disablePostgrest | bool | `true` |  |
| config-db.logLevel | string | `"{{.Values.global.logLevel}}"` |  |
| db.conf.checkpoint_completion_target | float | `0.9` |  |
| db.conf.effective_cache_size | string | `"3GB"` |  |
| db.conf.effective_io_concurrency | int | `200` |  |
| db.conf.extra_float_digits | int | `0` |  |
| db.conf.lc_messages | string | `"C"` |  |
| db.conf.listen_addresses | string | `"*"` |  |
| db.conf.log_autovacuum_min_duration | int | `0` |  |
| db.conf.log_checkpoints | string | `"on"` |  |
| db.conf.log_connections | string | `"on"` |  |
| db.conf.log_destination | string | `"csvlog"` |  |
| db.conf.log_disconnections | string | `"on"` |  |
| db.conf.log_filename | string | `"postgresql-%d.log"` |  |
| db.conf.log_lock_waits | string | `"on"` |  |
| db.conf.log_min_duration_statement | string | `"10s"` |  |
| db.conf.log_rotation_age | string | `"1d"` |  |
| db.conf.log_rotation_size | string | `"100MB"` |  |
| db.conf.log_temp_files | int | `0` |  |
| db.conf.log_timezone | string | `"UTC"` |  |
| db.conf.log_truncate_on_rotation | string | `"on"` |  |
| db.conf.logging_collector | string | `"on"` |  |
| db.conf.maintenance_work_mem | string | `"205MB"` |  |
| db.conf.max_connections | int | `100` |  |
| db.conf.max_parallel_workers | int | `2` |  |
| db.conf.max_parallel_workers_per_gather | int | `2` |  |
| db.conf.max_wal_size | string | `"3GB"` |  |
| db.conf.max_worker_processes | int | `8` |  |
| db.conf.min_wal_size | string | `"2GB"` |  |
| db.conf.password_encryption | string | `"scram-sha-256"` |  |
| db.conf.random_page_cost | float | `1.1` |  |
| db.conf.shared_buffers | string | `"1GB"` |  |
| db.conf.ssl | string | `"off"` |  |
| db.conf.timezone | string | `"UTC"` |  |
| db.conf.wal_buffers | int | `-1` |  |
| db.conf.work_mem | string | `"10MB"` |  |
| db.create | bool | `true` |  |
| db.jwtSecretKeyRef.key | string | `"PGRST_JWT_SECRET"` |  |
| db.jwtSecretKeyRef.name | string | `"incident-commander-postgrest-jwt"` |  |
| db.pganalyze.enabled | bool | `false` |  |
| db.pganalyze.secretName | string | `"pganalyze"` |  |
| db.resources.requests.memory | string | `"2Gi"` |  |
| db.secretKeyRef.key | string | `"DB_URL"` |  |
| db.secretKeyRef.name | string | `"incident-commander-postgres"` |  |
| db.shmVolume | string | `"256Mi"` |  |
| db.storage | string | `"20Gi"` |  |
| db.storageClass | string | `""` |  |
| global.imagePrefix | string | `"flanksource"` |  |
| global.imageRegistry | string | `"public.ecr.aws"` |  |
| global.labels | object | `{}` |  |
| global.logLevel | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"docker.io/flanksource/incident-commander"` |  |
| image.tag | string | `"v0.0.1313"` |  |
| jsonLogs | bool | `true` |  |
| logLevel | string | `"{{.Values.global.logLevel}}"` |  |
| pushTelemetry.enabled | bool | `false` |  |
| pushTelemetry.isAgent | bool | `true` |  |
| pushTelemetry.playbooks | bool | `false` |  |
| pushTelemetry.pushLocation.url | string | `"https://telemetry.app.flanksource.com/push/topology"` |  |
| resources.limits.cpu | string | `"500m"` |  |
| resources.limits.memory | string | `"1024Mi"` |  |
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"768Mi"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.name | string | `"mission-control-sa"` |  |
| serviceAccount.rbac.clusterRole | bool | `true` |  |
| serviceAccount.rbac.configmaps | bool | `true` |  |
| serviceAccount.rbac.exec | bool | `true` |  |
| serviceAccount.rbac.podRun | bool | `true` |  |
| serviceAccount.rbac.readAll | bool | `true` |  |
| serviceAccount.rbac.secrets | bool | `true` |  |
| serviceAccount.rbac.tokenRequest | bool | `true` |  |
| upstream.agentName | string | `""` |  |
| upstream.createSecret | bool | `true` |  |
| upstream.host | string | `""` |  |
| upstream.password | string | `""` |  |
| upstream.secretName | string | `"upstream"` |  |
| upstream.username | string | `""` |  |

