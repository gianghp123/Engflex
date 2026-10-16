# One-click `type = "minio"` does not exist in this Coolify version's service
# template bundle (359 keys, no minio — API answers 404 Invalid service type),
# so the service is created from Coolify's own canonical minio compose
# (templates/compose/minio.yaml `services:` block, comments stripped) via
# docker_compose_raw. Image comes from the loopback registry
# (127.0.0.1:5000/minio:latest — quay.io/Hub are not anonymously pullable;
# see the SDD ledger). The service joins the shared `coolify` network with
# the `minio` alias so same-project containers reach it at http://minio:9000
# (bare resource names do not resolve on this network). NOTE: heredoc
# `$VAR`/`{{project.VAR}}` stay literal in HCL (only `${` interpolates), so
# the references reach Coolify intact — and deliberately so: values live in
# project-level env (Coolify pre-creates per-resource keys for every $VAR
# ref at service birth; any TF-declared env resource for the same key then
# 409s forever). No per-resource coolify_environment_variable may exist for
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
  urls               = [{ name = "minio", url = "http://${var.minio_host}" }]
  # Kept true: Coolify answers 400 to PATCHing this service, so the flag
  # cannot be flipped on a live service. Fresh deploys therefore still
  # instant-boot with placeholder rows first; the bulk upsert adopts the
  # real values right after.
  instant_deploy = true
}

# Bulk upsert owns service env values (same rule as apps: no per-key
# resources). Rollout owner is the action below: any env change replaces it
# and Coolify restarts the service. On birth it fires `start` alongside the
# instant self-boot — start is idempotent (one healthy container per birth).
resource "coolify_envs_bulk" "this" {
  resource_type = "service"
  resource_uuid = coolify_service.this.uuid
  variables     = var.environment_variables
}

resource "coolify_resource_action" "boot" {
  action        = "start"
  resource_type = "service"
  resource_uuid = coolify_service.this.uuid
  triggers      = { env = sha256(jsonencode(var.environment_variables)) }
  depends_on    = [coolify_service.this, coolify_envs_bulk.this]
}
