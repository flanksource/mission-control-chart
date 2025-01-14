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

.PHONY: values.schema.json
values.schema.json: .bin/helm-schema
	cd chart && ../.bin/helm-schema -r -f values.yaml && cd -

.bin/helm-schema:
	test -s $(LOCALBIN)/helm-schema  || \
	GOBIN=$(LOCALBIN) go install github.com/dadav/helm-schema/cmd/helm-schema@latest
