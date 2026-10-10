SHELL := /bin/bash
.PHONY: plan deploy deploy-seed pronounce-artifact

plan:
	terraform -chdir=infrastructure/environments/selfhost plan

deploy:
	bash infrastructure/scripts/deploy.sh

deploy-seed:
	bash infrastructure/scripts/seed.sh

pronounce-artifact:
	@bash infrastructure/scripts/ensure-pronounce-image.sh $(ARGS)
