# One-click `type = "minio"` does not exist in this Coolify version's service
# template bundle (359 keys, no minio — API answers 404 Invalid service type),
# so the service is created from Coolify's own canonical minio compose
# (templates/compose/minio.yaml `services:` block, comments stripped) via
# docker_compose_raw. Image comes from the loopback registry
# (127.0.0.1:5000/minio:latest — quay.io/Hub are not anonymously pullable;
# see the SDD ledger). The service joins the shared `coolify` network with
# the `minio` alias so same-project containers reach it at http://minio:9000
# (resource-name DNS was disproven live on this instance). NOTE: heredoc
# `$VAR` stays literal in HCL (only `${` interpolates), so the compose
# `$SERVICE_USER_MINIO`-style references reach Coolify intact.
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
  instant_deploy     = true
}

resource "coolify_environment_variable" "root_user" {
  service_uuid = coolify_service.this.uuid
  key          = "SERVICE_USER_MINIO"
  value        = var.root_user
}

resource "coolify_environment_variable" "root_password" {
  service_uuid = coolify_service.this.uuid
  key          = "SERVICE_PASSWORD_MINIO"
  value        = var.root_password
}

resource "coolify_environment_variable" "server_url" {
  service_uuid = coolify_service.this.uuid
  key          = "MINIO_SERVER_URL"
  value        = "https://${var.minio_host}"
}

resource "coolify_environment_variable" "browser_redirect" {
  service_uuid = coolify_service.this.uuid
  key          = "MINIO_BROWSER_REDIRECT_URL"
  value        = "https://${var.minio_host}"
}
