# mission-control

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: 0.0.10](https://img.shields.io/badge/AppVersion-0.0.10-informational?style=flat-square)

A Helm chart for flanksource mission control

## Maintainers

| Name | Email | Url |
| ---- | ------ | --- |
| Flanksource |  |  |

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://flanksource.github.io/charts | apm-hub | >= 0.0.20 |
| https://flanksource.github.io/charts | canary-checker | 1.1.2-beta.115 |
| https://flanksource.github.io/charts | config-db | 0.0.989 |
| https://flanksource.github.io/charts | flanksource-ui | 1.4.36 |
| https://k8s.ory.sh/helm/charts | kratos | 0.32.0 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| adminPassword.secretKeyRef.create | bool | `true` |  |
| adminPassword.secretKeyRef.key | string | `"password"` |  |
| adminPassword.secretKeyRef.name | string | `"mission-control-admin-password"` |  |
| apm-hub.db.enabled | bool | `false` |  |
| apm-hub.db.secretKeyRef.create | bool | `false` |  |
| apm-hub.db.secretKeyRef.key | string | `"DB_URL"` |  |
| apm-hub.db.secretKeyRef.name | string | `"incident-commander-postgres"` |  |
| apm-hub.enabled | bool | `false` |  |
| artifactConnection | string | `""` | artifact connection string |
| authProvider | string | `"kratos"` |  |
| canary-checker.db.external.create | bool | `false` |  |
| canary-checker.db.external.enabled | bool | `true` |  |
| canary-checker.db.external.secretKeyRef.key | string | `"DB_URL"` |  |
| canary-checker.db.external.secretKeyRef.name | string | `"incident-commander-postgres"` |  |
| canary-checker.db.runMigrations | bool | `false` |  |
| canary-checker.disablePostgrest | bool | `true` |  |
| canary-checker.flanksource-ui.enabled | bool | `false` |  |
| canary-checker.image.type | string | `"full"` |  |
| canary-checker.logLevel | string | `"{{.Values.global.logLevel}}"` |  |
| canary-checker.otel.collector | string | `"{{ .Values.global.otel.collector }}"` |  |
| canary-checker.otel.labels | string | `"{{ .Values.global.otel.labels }}"` |  |
| cleanupResourcesOnDelete | bool | `false` |  |
| clerkJWKSURL | string | `""` |  |
| clerkOrgID | string | `""` |  |
| config-db.db.embedded.persist | bool | `false` |  |
| config-db.db.external.enabled | bool | `true` |  |
| config-db.db.external.secretKeyRef.key | string | `"DB_URL"` |  |
| config-db.db.external.secretKeyRef.name | string | `"incident-commander-postgres"` |  |
| config-db.db.runMigrations | bool | `false` |  |
| config-db.disablePostgrest | bool | `true` |  |
| config-db.logLevel | string | `"{{.Values.global.logLevel}}"` |  |
| config-db.otel.collector | string | `"{{ .Values.global.otel.collector }}"` |  |
| config-db.otel.labels | string | `"{{ .Values.global.otel.labels }}"` |  |
| db.conf.db_user_namespace | string | `"off"` |  |
| db.conf.effective_cache_size | string | `"3GB"` |  |
| db.conf.effective_io_concurrency | int | `200` |  |
| db.conf.extra_float_digits | int | `0` |  |
| db.conf.log_autovacuum_min_duration | int | `0` |  |
| db.conf.log_connections | string | `"on"` |  |
| db.conf.log_destination | string | `"stderr"` |  |
| db.conf.log_directory | string | `"/var/log/postgresql"` |  |
| db.conf.log_file_mode | int | `420` |  |
| db.conf.log_filename | string | `"postgresql-%d.log"` |  |
| db.conf.log_line_prefix | string | `"%m [%p] %q[user=%u,db=%d,app=%a] "` |  |
| db.conf.log_lock_waits | string | `"on"` |  |
| db.conf.log_min_duration_statement | string | `"1s"` |  |
| db.conf.log_rotation_age | string | `"1d"` |  |
| db.conf.log_rotation_size | string | `"100MB"` |  |
| db.conf.log_statement | string | `"all"` |  |
| db.conf.log_temp_files | int | `0` |  |
| db.conf.log_timezone | string | `"UTC"` |  |
| db.conf.log_truncate_on_rotation | string | `"on"` |  |
| db.conf.logging_collector | string | `"on"` |  |
| db.conf.maintenance_work_mem | string | `"256MB"` |  |
| db.conf.max_connections | int | `200` |  |
| db.conf.max_wal_size | string | `"4GB"` |  |
| db.conf.password_encryption | string | `"scram-sha-256"` |  |
| db.conf.shared_buffers | string | `"1GB"` |  |
| db.conf.ssl | string | `"off"` |  |
| db.conf.timezone | string | `"UTC"` |  |
| db.conf.wal_buffers | string | `"16MB"` |  |
| db.conf.work_mem | string | `"10MB"` |  |
| db.create | bool | `true` |  |
| db.jwtSecretKeyRef.key | string | `"PGRST_JWT_SECRET"` |  |
| db.jwtSecretKeyRef.name | string | `"incident-commander-postgrest-jwt"` |  |
| db.pganalyze.enabled | bool | `false` |  |
| db.pganalyze.secretName | string | `"pganalyze"` |  |
| db.pganalyze.systemID | string | `"mission-control"` |  |
| db.resources.requests.memory | string | `"4Gi"` |  |
| db.secretKeyRef.key | string | `"DB_URL"` |  |
| db.secretKeyRef.name | string | `"incident-commander-postgres"` |  |
| db.shmVolume | string | `"256Mi"` |  |
| db.storage | string | `"20Gi"` |  |
| db.storageClass | string | `""` |  |
| externalPostgrest.dbAnonRole | string | `"postgrest_anon"` |  |
| externalPostgrest.dbScema | string | `"public"` |  |
| externalPostgrest.enable | bool | `true` |  |
| externalPostgrest.imageName | string | `""` |  |
| externalPostgrest.logLevel | string | `"info"` |  |
| externalPostgrest.maxRows | int | `2000` |  |
| externalPostgrest.tag | string | `"v10.2.0"` |  |
| extraArgs | object | `{}` |  |
| flanksource-ui.backendURL | string | `"http://mission-control:8080"` |  |
| flanksource-ui.enabled | bool | `true` |  |
| flanksource-ui.fullnameOverride | string | `"incident-manager-ui"` |  |
| flanksource-ui.ingress.enabled | bool | `true` |  |
| flanksource-ui.ingress.host | string | `"{{.Values.global.ui.host}}"` |  |
| flanksource-ui.ingress.tls[0].hosts[0] | string | `"{{.Values.global.ui.host}}"` |  |
| flanksource-ui.ingress.tls[0].secretName | string | `"{{.Values.global.ui.tlsSecretName}}"` |  |
| flanksource-ui.nameOverride | string | `"incident-manager-ui"` |  |
| flanksource-ui.oryKratosURL | string | `"http://{{.Values.global.ui.host}}/api/.ory"` |  |
| global.api.host | string | `"mission-control-ui.local/api"` |  |
| global.api.tlsSecretName | string | `""` |  |
| global.db.connectionPooler.enabled | bool | `false` |  |
| global.db.connectionPooler.extraContainers | string | `""` |  |
| global.db.connectionPooler.image | string | `"bitnami/pgbouncer:1.22.0"` |  |
| global.db.connectionPooler.secretKeyRef.key | string | `"DB_URL"` |  |
| global.db.connectionPooler.secretKeyRef.name | string | `"mission-control-connection-pooler"` |  |
| global.db.connectionPooler.serviceAccount.annotations | object | `{}` |  |
| global.imagePrefix | string | `"flanksource"` |  |
| global.imageRegistry | string | `"public.ecr.aws"` |  |
| global.labels | object | `{}` |  |
| global.logLevel | string | `""` |  |
| global.otel.collector | string | `""` |  |
| global.otel.labels | string | `""` |  |
| global.serviceMonitor.enabled | bool | `false` |  |
| global.serviceMonitor.labels | object | `{}` |  |
| global.ui.host | string | `"mission-control-ui.local"` |  |
| global.ui.tlsSecretName | string | `"mission-control-ui-tls"` |  |
| grafana.scrapeMetricsDashboard.enabled | bool | `false` |  |
| grafana.scrapeMetricsDashboard.labels.grafana_dashboard | string | `"1"` |  |
| identityRoleMapper.configMap.key | string | `""` |  |
| identityRoleMapper.configMap.mountPath | string | `"/etc/identity-role-mapper"` |  |
| identityRoleMapper.configMap.name | string | `""` |  |
| identityRoleMapper.script | string | `""` |  |
| image.pullPolicy | string | `"IfNotPresent"` |  |
| image.repository | string | `"docker.io/flanksource/incident-commander"` |  |
| image.tag | string | `"v0.0.1313"` |  |
| impersonationRole.createNamespaces | bool | `true` |  |
| impersonationRole.namespaces[0] | string | `"default"` |  |
| ingress.annotations."kubernetes.io/ingress.class" | string | `"nginx"` |  |
| ingress.enabled | bool | `false` |  |
| ingress.host | string | `"{{.Values.global.api.host}}"` |  |
| ingress.tls[0].hosts[0] | string | `"{{.Values.global.api.host}}"` |  |
| ingress.tls[0].secretName | string | `"{{.Values.global.api.tlsSecretName}}"` |  |
| jsonLogs | bool | `true` |  |
| kmsConnection | string | `""` | kms connection string |
| kratos.automigration.customArgs[0] | string | `"migrate"` |  |
| kratos.automigration.customArgs[1] | string | `"sql"` |  |
| kratos.automigration.customArgs[2] | string | `"-e"` |  |
| kratos.automigration.customArgs[3] | string | `"--yes"` |  |
| kratos.automigration.customArgs[4] | string | `"--config"` |  |
| kratos.automigration.customArgs[5] | string | `"/etc/custom/config/kratos.yaml"` |  |
| kratos.configmap.hashSumEnabled | bool | `false` |  |
| kratos.courier.enabled | bool | `false` |  |
| kratos.deployment.extraArgs[0] | string | `"--watch-courier"` |  |
| kratos.deployment.extraArgs[1] | string | `"--config"` |  |
| kratos.deployment.extraArgs[2] | string | `"/etc/custom/config/kratos.yaml"` |  |
| kratos.deployment.extraVolumeMounts[0].mountPath | string | `"/etc/custom/config"` |  |
| kratos.deployment.extraVolumeMounts[0].name | string | `"kratos-custom-config-volume"` |  |
| kratos.deployment.extraVolumeMounts[0].readOnly | bool | `true` |  |
| kratos.deployment.extraVolumes[0].configMap.name | string | `"mission-control-kratos-config"` |  |
| kratos.deployment.extraVolumes[0].name | string | `"kratos-custom-config-volume"` |  |
| kratos.enabled | bool | `true` |  |
| kratos.fullnameOverride | string | `"kratos"` |  |
| kratos.image.repository | string | `"public.ecr.aws/k4y9r6y5/kratos"` |  |
| kratos.ingress.public.enabled | bool | `false` |  |
| kratos.kratos.automigration.enabled | bool | `true` |  |
| kratos.kratos.automigration.type | string | `"initContainer"` |  |
| kratos.kratos.config.courier.smtp.connection_uri | string | `"smtp://wrong-url"` |  |
| kratos.kratos.config.log.level | string | `"warning"` |  |
| kratos.kratos.config.secrets.default[0] | string | `"yet another secret"` |  |
| kratos.kratos.config.secrets.default[1] | string | `"lorem ipsum dolores"` |  |
| kratos.kratos.config.secrets.default[2] | string | `"just a random a string secret"` |  |
| kratos.kratos.config.session.lifespan | string | `"336h"` |  |
| kratos.secret.enabled | bool | `false` |  |
| logLevel | string | `"{{.Values.global.logLevel}}"` |  |
| nameOverride | string | `""` |  |
| otel.collector | string | `"{{.Values.global.otel.collector}}"` |  |
| otel.labels | string | `"{{ .Values.global.otel.labels }}"` |  |
| otel.serviceName | string | `"mission-control"` |  |
| permissions.components | bool | `false` | when enabled, services must have explicit permissions to read components otherwise, system automatically has permission to read all components. |
| permissions.configs | bool | `false` | when enabled, services must have explicit permissions to read configs otherwise, system automatically has permission to read all configs. |
| permissions.connections | bool | `false` | when enabled, users & services must have explicit permissions to run connections otherwise, editors automatically have permission to run connections. |
| permissions.playbooks | bool | `false` | when enabled, users & services must have explicit permissions to run playbooks otherwise, editors automatically have permission to run playbooks. |
| properties."incidents.disable" | bool | `true` |  |
| properties."logs.disable" | bool | `true` |  |
| replicas | int | `1` |  |
| resources.limits.cpu | string | `"500m"` |  |
| resources.limits.memory | string | `"1024Mi"` |  |
| resources.requests.cpu | string | `"100m"` |  |
| resources.requests.memory | string | `"768Mi"` |  |
| serviceAccount.annotations | object | `{}` |  |
| serviceAccount.name | string | `"mission-control-sa"` |  |
| serviceAccount.rbac.clusterAdmin | bool | `false` |  |
| serviceAccount.rbac.clusterRole | bool | `true` |  |
| serviceAccount.rbac.configmaps | bool | `true` |  |
| serviceAccount.rbac.exec | bool | `true` |  |
| serviceAccount.rbac.extra | list | `[]` |  |
| serviceAccount.rbac.impersonate | bool | `false` |  |
| serviceAccount.rbac.podRun | bool | `true` |  |
| serviceAccount.rbac.readAll | bool | `true` |  |
| serviceAccount.rbac.secrets | bool | `true` |  |
| serviceAccount.rbac.tokenRequest | bool | `true` |  |
| serviceMonitor.enabled | bool | `false` |  |
| serviceMonitor.labels | object | `{}` |  |
| smtp.secretRef.name | string | `"incident-commander-smtp"` |  |
| upstream_push | object | `{}` |  |

