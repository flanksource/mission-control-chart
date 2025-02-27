LOCALBIN ?= $(shell pwd)/.bin

$(LOCALBIN):
	mkdir -p .bin

.PHONY: chart-local
chart-local:
	helm dependency update ./chart
	helm template -f ./chart/values.yaml mission-control ./chart

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
		'reduce ["apm-hub", "canary-checker", "config-db","flanksource-ui", "kratos"][] as $$key (.; .properties[$$key] = $$JSON_CONTENT) | .additionalProperties = true' \
		chart/temp-schema.json > chart/values.schema.json
	rm chart/temp-schema.json

.bin/helm-schema:
	test -s $(LOCALBIN)/helm-schema  || \
	GOBIN=$(LOCALBIN) go install github.com/dadav/helm-schema/cmd/helm-schema@latest

.PHONY: README.md
README.md: .bin/helm-docs
	.bin/helm-docs -t README.md.tpl

.bin/helm-docs:
	test -s $(LOCALBIN)/helm-docs  || \
	GOBIN=$(LOCALBIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest