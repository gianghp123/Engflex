SHELL := /bin/bash
.PHONY: plan deploy deploy-seed

plan:
	terraform -chdir=infrastructure/environments/selfhost plan

deploy:
	bash infrastructure/scripts/deploy.sh

deploy-seed:
	bash infrastructure/scripts/seed.sh
