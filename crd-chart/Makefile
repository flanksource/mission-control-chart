download:
	rm -rf templates/
	mkdir -p templates
	cp ../chart/crds/* templates/
download:
	rm -rf templates/
	mkdir -p templates
	cp ../chart/crds/* templates/

	for i in Canary Component Topology; do \
			wget -O templates/canary-checker.flanksource.com_$$i.yaml https://github.com/flanksource/canary-checker/raw/refs/heads/master/config/deploy/$$i.yml; \
	done


	for i in scrapeconfigs scrapeplugins; do \
		wget -O templates/configs.flanksource.com_$$i.yaml https://github.com/flanksource/config-db/raw/main/chart/crds/configs.flanksource.com_$$i.yaml; \
	done


	for i in connections incidentrules notifications notificationsilences permissions permissiongroups playbooks; do \
		wget -O templates/mission-control.flanksource.com_$$i.yaml \
			 https://github.com/flanksource/mission-control/raw/main/config/crds/mission-control.flanksource.com_$$i.yaml;  \
	done

