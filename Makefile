LOCALBIN ?= $(shell pwd)/.bin

$(LOCALBIN):
	mkdir -p .bin

.PHONY: chart-local
chart-local:
	helm dependency update ./chart
	helm template -f ./chart/values.yaml mission-control ./chart

.PHONY: test-templates
test-templates:
	helm dependency update ./chart
	helm template --dry-run -f ./chart/values.yaml mission-control ./chart > /dev/null && echo "Templates are valid"

.PHONY: chart
chart:
	helm dependency build ./chart
	helm package ./chart

.PHONY: agent-chart
agent-chart:
	helm dependency build ./agent-chart
	helm package ./agent-chart

.PHONY: crd-chart
crd-chart:
	helm package ./crd-chart

# We have to update schema for chart dependencies to freely modify its properties
.PHONY: values.schema.json
values.schema.json: .bin/helm-schema
	cd chart && ../.bin/helm-schema -r -f values.yaml -o temp-schema.json && cd -
	jq  --argjson JSON_CONTENT '{"additionalProperties": true,"required": [],"type": "object"}' \
		'reduce ["apm-hub", "canary-checker", "config-db","flanksource-ui", "kratos", "mission-control-kubernetes-view"][] as $$key (.; .properties[$$key] = $$JSON_CONTENT) | .additionalProperties = true' \
		chart/temp-schema.json > chart/values.schema.json
	rm chart/temp-schema.json
	cd agent-chart && ../.bin/helm-schema -r -f values.yaml -o temp-schema.json && cd -
	jq  --argjson JSON_CONTENT '{"additionalProperties": true,"required": [],"type": "object"}' \
		'reduce ["apm-hub", "canary-checker", "config-db","flanksource-ui", "kratos", "mission-control-kubernetes-view", "pushTelemetry"][] as $$key (.; .properties[$$key] = $$JSON_CONTENT) | .additionalProperties = true' \
		agent-chart/temp-schema.json > agent-chart/values.schema.json
	rm agent-chart/temp-schema.json

.bin/helm-schema:
	test -s $(LOCALBIN)/helm-schema  || \
	GOBIN=$(LOCALBIN) go install github.com/dadav/helm-schema/cmd/helm-schema@latest

.PHONY: README.md
README.md: .bin/helm-docs
	.bin/helm-docs -t README.md.tpl

.bin/helm-docs:
	test -s $(LOCALBIN)/helm-docs  || \
	GOBIN=$(LOCALBIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest

.bin/ct: $(LOCALBIN)
	@if [ ! -f $(LOCALBIN)/ct ]; then \
		echo "Downloading chart-testing (ct) to .bin/..."; \
		OS=$$(uname -s | tr '[:upper:]' '[:lower:]'); \
		ARCH=$$(uname -m); \
		if [ "$$ARCH" = "x86_64" ]; then ARCH="amd64"; fi; \
		if [ "$$ARCH" = "aarch64" ]; then ARCH="arm64"; fi; \
		CT_VERSION="v3.11.0"; \
		URL="https://github.com/helm/chart-testing/releases/download/$${CT_VERSION}/chart-testing_$${CT_VERSION#v}_$${OS}_$${ARCH}.tar.gz"; \
		curl -sSL "$$URL" | tar -xz -C $(LOCALBIN) ct etc; \
		chmod +x $(LOCALBIN)/ct; \
	fi

.PHONY: lint
lint: .bin/ct
	$(LOCALBIN)/ct lint --charts chart --chart-yaml-schema $(LOCALBIN)/etc/chart_schema.yaml --lint-conf $(LOCALBIN)/etc/lintconf.yaml
