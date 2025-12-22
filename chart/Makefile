
LOCALBIN ?= $(shell pwd)/.bin

$(LOCALBIN):
	mkdir -p .bin

.PHONY: values.schema.json
values.schema.json: .bin/helm-schema
	.bin/helm-schema -r -n -f values.yaml
	jq 'del(.properties.kratos.properties)'  values.schema.json | sponge   values.schema.json
	jq 'del(.properties.kratos.required)' values.schema.json | sponge   values.schema.json
	jq 'del(.properties."flanksource-ui".properties)' values.schema.json | sponge   values.schema.json
	jq 'del(.properties."flanksource-ui"."$$ref")' values.schema.json | sponge   values.schema.json



.PHONY: chart
chart: values.schema.deref.json  README.md


.PHONY: README.md
README.md: .bin/helm-docs
	.bin/helm-docs   -t README.md.tpl


.bin/helm-docs:
	test -s $(LOCALBIN)/helm-docs  || \
	GOBIN=$(LOCALBIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@latest

.bin/helm-schema:
	test -s $(LOCALBIN)/helm-schema  || \
	GOBIN=$(LOCALBIN) go install github.com/dadav/helm-schema/cmd/helm-schema@0.18.1


.phony: values.schema.deref.json
values.schema.deref.json: values.schema.json
	npm init --init-module chart --init-version 1.0.0 -y --force
	npm i @apidevtools/json-schema-ref-parser --save-dev
	npm i @digitak/esrun  --save-dev
	npx esrun  deref.ts | jq > values.schema.deref.json


.PHONY: lint
lint: chart
	ct lint --charts .
