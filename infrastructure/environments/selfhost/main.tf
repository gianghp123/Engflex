locals {
  ip_dashed  = replace(var.server_ip, ".", "-")
  api_host   = "engflex-api.${local.ip_dashed}.sslip.io"
  web_host   = "engflex-web.${local.ip_dashed}.sslip.io"
  minio_host = "engflex-minio.${local.ip_dashed}.sslip.io"

  # internal container traffic stays http by intent.
  api_url   = "https://${local.api_host}"
  web_url   = "https://${local.web_host}"
  minio_url = "https://${local.minio_host}"
}

resource "coolify_project" "engflex" {
  name        = "engflex"
  description = "Managed by Terraform - see infrastructure/environments/selfhost"
}

module "postgres" {
  source       = "../../modules/postgres"
  name         = "engflex-postgres"
  project_uuid = coolify_project.engflex.uuid
  server_uuid  = var.server_uuid
  db_name      = var.db_name
  db_user      = var.db_user
  db_password  = var.db_password
  public_port  = var.db_public_port
}

module "redis" {
  source         = "../../modules/redis"
  name           = "engflex-redis"
  project_uuid   = coolify_project.engflex.uuid
  server_uuid    = var.server_uuid
  redis_password = var.redis_password
}

module "api_app" {
  source         = "../../modules/application"
  name           = "engflex-api"
  project_uuid   = coolify_project.engflex.uuid
  server_uuid    = var.server_uuid
  git_repository = var.git_repository

  # Repo-root context: dockerfile_location is repo-relative.
  base_directory      = "/"
  dockerfile_location = "/apps/api/Dockerfile"
  ports_exposes       = "8000"

  # Stable internal DNS on the shared coolify network (verified live:
  # containers are otherwise aliased only by versioned names).
  network_aliases = "engflex-api"

  domains           = [local.api_url]
  health_check_path = "/health"
  instant_deploy    = false

  environment_variables = {}
}

module "web_app" {
  source         = "../../modules/application"
  name           = "engflex-web"
  project_uuid   = coolify_project.engflex.uuid
  server_uuid    = var.server_uuid
  git_repository = var.git_repository

  base_directory      = "/"
  dockerfile_location = "/apps/web/Dockerfile"
  ports_exposes       = "3000"

  domains           = concat([local.web_url], var.cloudflare_enabled ? ["https://${module.tunnel[0].hostname}"] : [])
  health_check_path = "/"
  instant_deploy    = false

  # Tunnel (web-only front door): the tunnel delivers plain HTTP to the
  # proxy, so the https redirect must stay off while enabled — otherwise
  # origin 301s loop into TOO_MANY_REDIRECTS (Coolify guide step 5).
  # Direct https:// sslip access keeps working (TLS still terminates).
  is_force_https_enabled = !var.cloudflare_enabled
  environment_variables  = {}
}

module "voice_app" {
  source                  = "../../modules/application"
  name                    = "engflex-voice"
  project_uuid            = coolify_project.engflex.uuid
  server_uuid             = var.server_uuid
  git_repository          = var.git_repository
  build_pack              = "dockercompose"
  base_directory          = "/"
  docker_compose_location = "/infrastructure/voice/compose.yaml"
  ports_exposes           = "7860"
  domains                 = []
  health_check_path       = "/status"
  instant_deploy          = false
  # Bulk-managed env (see ledger 2026-10-10): TF owns no per-key rows;
  # values go in via PATCH bulk, deploy.sh triggers the rollout.
  environment_variables = {}
}

module "pronounce_app" {
  source              = "../../modules/application"
  name                = "engflex-pronounce"
  project_uuid        = coolify_project.engflex.uuid
  server_uuid         = var.server_uuid
  git_repository      = var.git_repository
  base_directory      = "/"
  dockerfile_location = "/infrastructure/pronounce/Dockerfile"
  ports_exposes       = "8000"
  domains             = []
  health_check_path   = "/health"
  instant_deploy      = false
  # Stable internal DNS on the shared coolify network (same rule as api).
  network_aliases       = "engflex-pronounce"
  environment_variables = {}
}

module "minio" {
  source        = "../../modules/minio"
  name          = "engflex-minio"
  project_uuid  = coolify_project.engflex.uuid
  server_uuid   = var.server_uuid
  minio_host    = local.minio_host
  root_user     = var.minio_root_user
  root_password = var.minio_root_password
}

# Cloudflare Tunnel in front of the Coolify proxy (all-resource pattern).
# Disabled by default — set cloudflare_enabled = true plus zone vars once
# DNS is on Cloudflare. deploy.sh triggers the agent's first boot AFTER
# the token env exists (instant_deploy = false in the module).
module "tunnel" {
  count                  = var.cloudflare_enabled ? 1 : 0
  source                 = "../../modules/tunnel"
  name                   = "engflex-cloudflared"
  project_uuid           = coolify_project.engflex.uuid
  server_uuid            = var.server_uuid
  instant_deploy         = false
  cloudflare_tunnel_name = var.cloudflare_tunnel_name
  cloudflare_account_id  = var.cloudflare_account_id
  cloudflare_email       = var.cloudflare_email
  cloudflare_token       = var.cloudflare_token
  cloudflare_zone_id     = var.cloudflare_zone_id
  cloudflare_zone        = var.cloudflare_zone
  app_name               = var.cloudflare_app_name
}

# Sentry projects + DSNs for the three instrumented services. The DSNs feed
# the app env maps above; the -production suffix is this deployment's
# environment — a future staging env passes its own environment value
# without renaming existing projects.
module "sentry" {
  source       = "../../modules/sentry"
  organization = var.sentry_organization
  team         = var.sentry_team
  environment  = "production"

  projects = {
    api   = { name = "EngFlex API", slug = "engflex-api", platform = "go" }
    web   = { name = "EngFlex Web", slug = "engflex-web", platform = "javascript-react" }
    voice = { name = "EngFlex Voice", slug = "engflex-voice", platform = "python" }
  }
}
