download:
	rm -rf templates/
	mkdir -p templates
	cp ../chart/crds/* templates/
	wget -O templates/configs.flanksource.com_scrapeconfigs.yaml https://github.com/flanksource/config-db/raw/main/chart/crds/configs.flanksource.com_scrapeconfigs.yaml
	wget -O templates/canary-checker.flanksource.com.yaml https://github.com/flanksource/canary-checker/raw/master/config/deploy/crd.yaml

