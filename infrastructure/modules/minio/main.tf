# One-click `type = "minio"` does not exist in this Coolify version's service
# template bundle (359 keys, no minio — API answers 404 Invalid service type),
# so the service is created from Coolify's own canonical minio compose
# (templates/compose/minio.yaml `services:` block, comments stripped) via
# docker_compose_raw. Image comes from the loopback registry
# (127.0.0.1:5000/minio:latest — quay.io/Hub are not anonymously pullable;
# see the SDD ledger). The service joins the shared `coolify` network with
# the `minio` alias so same-project containers reach it at http://minio:9000
# (resource-name DNS was disproven live on this instance). NOTE: heredoc
# `$VAR`/`{{project.VAR}}` stay literal in HCL (only `${` interpolates), so
# the references reach Coolify intact — and deliberately so: values live in
# project-shared vars (../environments/selfhost/shared.tf), because Coolify
# pre-creates per-resource keys for every $VAR ref at service birth and any
# TF-declared env resource for the same key then 409s forever (post-mortem
# 2026-10-10). No per-resource coolify_environment_variable may exist for
# keys referenced in this compose. Ever.
resource "coolify_service" "this" {
  name               = var.name
  project_uuid       = var.project_uuid
  server_uuid        = var.server_uuid
  docker_compose_raw = <<-EOT
    services:
      minio:
        image: 127.0.0.1:5000/minio:latest
        command: server /data --console-address ":9001"
        environment:
          - MINIO_SERVER_URL=$MINIO_SERVER_URL
          - MINIO_BROWSER_REDIRECT_URL=$MINIO_BROWSER_REDIRECT_URL
          - MINIO_ROOT_USER=$SERVICE_USER_MINIO
          - MINIO_ROOT_PASSWORD=$SERVICE_PASSWORD_MINIO
        volumes:
          - minio-data:/data
        networks:
          coolify:
            aliases:
              - minio
        healthcheck:
          test: ["CMD", "mc", "ready", "local"]
          interval: 5s
          timeout: 20s
          retries: 10
    networks:
      coolify:
        external: true
  EOT
  urls               = [{ name = "minio", url = "https://${var.minio_host}" }]
  # Kept true: Coolify 4.4.6 answers 400 to PATCHing this service, so the
  # flag cannot be flipped on a live service (tried 2026-10-10). Fresh
  # deploys therefore still instant-boot with placeholder rows first —
  # adopt-once via adopt-service-envs.sh remains the flow (proven 4x).
  instant_deploy     = true
}

# EXPERIMENT 2026-10-10 (whole-stack no-trigger test): the 4 env resources
# (root_user, server_url, browser_redirect, root_password) are removed;
# values go in via service-bulk (birth placeholders 409 POSTs, bulk
# upserts). Restore from git when the verdict lands.

# NOTE (final rule 2026-10-10, voice prod incident): per-resource TF
# resources are PRIMARY for every key this compose references. A shared-var
# experiment ({{project.*}} refs + zero per-resource rows) booted with EMPTY
# creds in prod: auto-reborn NULL local rows shadow shared refs at deploy
# (local wins even when NULL; shared serves only with zero locals present).
# Fresh deploys still 409 on these two keys (platform pre-creates them) —
# adopt once via infrastructure/scripts/adopt-service-envs.sh, values then
# stick across redeploys (proven). shared.tf minio keys stay as spare.
