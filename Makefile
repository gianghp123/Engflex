SHELL := /bin/bash
.PHONY: plan deploy deploy-seed destroy pronounce-artifact

plan:
	terraform -chdir=infrastructure/environments/selfhost plan

deploy:
	bash infrastructure/scripts/deploy.sh

deploy-seed:
	bash infrastructure/scripts/seed.sh

# Destructive teardown of everything Terraform manages. Requires typing
# 'destroy'; terraform then runs with -auto-approve. Preview with
# `terraform -chdir=infrastructure/environments/selfhost plan -destroy`.
destroy:
	@echo "This destroys EVERYTHING managed by Terraform:"
	@echo "  Coolify project + Postgres (ALL DATA) + Redis + api/web/voice/pronounce apps"
	@echo "  + MinIO service (ALL MEDIA) + cloudflared agent + Cloudflare Tunnel/DNS/Access"
	@echo "  + Sentry team and engflex-*-production projects."
	@echo "Survives: git mirror, local registry, mkcert certs, Coolify itself."
	@read -r -p "Type 'destroy' to proceed: " confirm; \
	if [ "$$confirm" != "destroy" ]; then echo "aborted — nothing was destroyed"; exit 1; fi; \
	if terraform -chdir=infrastructure/environments/selfhost state list 2>/dev/null | grep -q 'module\.tunnel\['; then \
	  echo "== phase 1: destroy cloudflared agent first =="; \
	  echo "   (Cloudflare refuses tunnel deletion with live connections: code 1022)"; \
	  terraform -chdir=infrastructure/environments/selfhost destroy -auto-approve \
	    -target='module.tunnel[0].coolify_service.this' \
	    -target='module.tunnel[0].coolify_environment_variable.tunnel_token'; \
	  echo "   polling tunnel connections (10s interval, 300s max)..."; \
	  CF_TOKEN=$$(grep -oP '^cloudflare_token\s*=\s*"\K[^"]+' infrastructure/environments/selfhost/terraform.tfvars); \
	  CF_ACCT=$$(grep -oP '^cloudflare_account_id\s*=\s*"\K[^"]+' infrastructure/environments/selfhost/terraform.tfvars); \
	  TUNNEL_ID=$$(terraform -chdir=infrastructure/environments/selfhost output -raw tunnel_id); \
	  if [ -z "$$TUNNEL_ID" ]; then echo "   no tunnel id — skipping poll"; \
	  else i=0; while [ $$i -lt 30 ]; do \
	    N=$$(curl -fsS --max-time 10 -H "Authorization: Bearer $$CF_TOKEN" \
	      "https://api.cloudflare.com/client/v4/accounts/$$CF_ACCT/cfd_tunnel/$$TUNNEL_ID" \
	      | python3 -c "import json,sys; print(len(json.load(sys.stdin)['result']['connections']))" 2>/dev/null || echo "?"); \
	    if [ "$$N" = "0" ]; then echo "   tunnel drained — no active connections"; break; fi; \
	    echo "   ...still $$N connection(s), waiting 10s"; sleep 10; i=$$((i+1)); \
	  done; fi; \
	fi; \
	terraform -chdir=infrastructure/environments/selfhost destroy

pronounce-artifact:
	@bash infrastructure/scripts/ensure-pronounce-image.sh $(ARGS)
