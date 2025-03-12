download:
	rm -rf templates/
	mkdir -p templates
	cp ../chart/crds/* templates/

	wget -O templates/configs.flanksource.com_scrapeconfigs.yaml https://github.com/flanksource/config-db/raw/main/chart/crds/configs.flanksource.com_scrapeconfigs.yaml
	
	wget -O templates/canary-checker.flanksource.com_canary.yaml https://github.com/flanksource/canary-checker/raw/master/config/deploy/Canary.yml
	wget -O templates/canary-checker.flanksource.com_component.yaml https://github.com/flanksource/canary-checker/raw/master/config/deploy/Component.yml
	wget -O templates/canary-checker.flanksource.com_topology.yaml https://github.com/flanksource/canary-checker/raw/master/config/deploy/Topology.yml

	for i in connections incidentrules notifications playbooks; do \
		wget -O templates/mission-control.flanksource.com_$$i.yaml \
			 https://github.com/flanksource/mission-control/raw/main/config/crds/mission-control.flanksource.com_$$i.yaml;  \
	done

