locals {
  ip_dashed = replace(var.server_ip, ".", "-")

  # Single source for per-app identity: every name/port below has 2+
  # consumers (its module block + cross-app service URLs). URLs compose from
  # parts — never retyped — so evaluated values stay byte-identical.
  apps = {
    api       = { name = "engflex-api", port = "8000" }
    web       = { name = "engflex-web", port = "3000" }
    voice     = { name = "engflex-voice", port = "7860" }
    pronounce = { name = "engflex-pronounce", port = "8000" }
  }

  # Same for services (names consumed by hosts below + module blocks).
  services = {
    minio  = { name = "engflex-minio" }
    tunnel = { name = "engflex-cloudflared" }
  }

  api_host   = "${local.apps.api.name}.${local.ip_dashed}.sslip.io"
  web_host   = "${local.apps.web.name}.${local.ip_dashed}.sslip.io"
  minio_host = "${local.services.minio.name}.${local.ip_dashed}.sslip.io"

  # Plain HTTP by intent (tunnel-only + localhost): no TLS on local routes.
  # Browsers reach the stack via the tunnel (real certs) or localhost;
  # direct-LAN https (mkcert) is not served.
  api_url   = "http://${local.api_host}"
  web_url   = "http://${local.web_host}"
  minio_url = "http://${local.minio_host}"

  # All remaining literals, grouped by domain. Call sites compose URLs from
  # these + the maps above — never retype a value (plan-diff guard proves
  # byte-identical on every regroup like this one).
  environment = "production"

  db    = { port = "5432", user = "engflex", name = "engflex", sslmode = "disable" }
  redis = { key_prefix = "engflex" }

  logging = { format = "json" }
  network = { trusted_proxies = "10.0.1.0/24,fdf6:8678:91f4::/64" }
  stun    = { url = "stun:stun.l.google.com:19302" }

  # Object storage: Docker-DNS alias + port + bucket shared by api/web maps.
  minio_net = { host = "minio", port = "9000", bucket = "lesson-audio" }

  # Engine caps (api-side VOICE_MAX_DURATION_SEC intentionally shares the
  # engine value — single source enforces they match).
  voice_limits         = { max_duration_sec = "300", idle_timeout_sec = "300" }
  pronounce_timeout_ms = "120000"

  # Request throttling budgets (apps/api RATE_LIMIT_*): ulule/limiter rates.
  # Opt-in by design — an unset rate means no limiter — so prod must declare
  # them explicitly rather than inherit a code default.
  rate_limit = { global = "300-M", session = "30-M" }

  sentry_projects = {
    api   = { name = "EngFlex API", slug = local.apps.api.name, platform = "go" }
    web   = { name = "EngFlex Web", slug = local.apps.web.name, platform = "javascript-react" }
    voice = { name = "EngFlex Voice", slug = local.apps.voice.name, platform = "python" }
  }
}

# Origin HEAD, read once at plan time for all four apps' commit triggers.
data "external" "origin_head" {
  program = ["sh", "-c", "echo \"{\\\"sha\\\":\\\"$(git ls-remote origin ${var.git_branch} | cut -f1)\\\"}\""]
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
  name           = local.apps.api.name
  project_uuid   = coolify_project.engflex.uuid
  server_uuid    = var.server_uuid
  git_repository = var.git_repository
  git_branch     = var.git_branch

  # Repo-root context: dockerfile_location is repo-relative.
  base_directory      = "/"
  dockerfile_location = "/apps/api/Dockerfile"
  ports_exposes       = local.apps.api.port

  # Stable internal DNS on the shared coolify network (verified live:
  # containers are otherwise aliased only by versioned names).
  network_aliases = "engflex-api"

  domains           = [local.api_url]
  health_check_path = "/health"

  # Plain-HTTP domains must not redirect (no TLS served locally anymore).
  is_force_https_enabled = false

  environment_variables = {
    DATABASE_URL               = module.postgres.internal_url
    DB_HOST                    = module.postgres.uuid
    DB_PORT                    = local.db.port
    DB_USER                    = local.db.user
    DB_PASSWORD                = var.db_password
    DB_NAME                    = local.db.name
    DB_SSLMODE                 = local.db.sslmode
    REDIS_URL                  = module.redis.internal_url
    REDIS_KEY_PREFIX           = local.redis.key_prefix
    MINIO_ENDPOINT             = "http://${local.minio_net.host}:${local.minio_net.port}"
    MINIO_ROOT_USER            = var.minio_root_user
    MINIO_ROOT_PASSWORD        = var.minio_root_password
    MINIO_BUCKET               = local.minio_net.bucket
    CLERK_SECRET_KEY           = var.clerk_secret_key
    CORS_ORIGINS               = join(",", [local.web_url, local.api_url])
    RATE_LIMIT_GLOBAL          = local.rate_limit.global
    RATE_LIMIT_SESSION         = local.rate_limit.session
    RATE_LIMIT_TRUSTED_PROXIES = local.network.trusted_proxies
    LLM_API_KEY                = var.llm_api_key
    LLM_BASE_URL               = var.llm_base_url
    LLM_MODEL                  = var.llm_model
    SENTRY_DSN                 = module.sentry.dsns["api"]
    VOICE_SERVICE_URL          = "http://${local.apps.voice.name}:${local.apps.voice.port}"
    VOICE_INTERNAL_SECRET      = var.voice_internal_secret
    VOICE_MAX_DURATION_SEC     = local.voice_limits.max_duration_sec
    PRONOUNCE_SERVICE_URL      = "http://${local.apps.pronounce.name}:${local.apps.pronounce.port}"
    PRONOUNCE_TIMEOUT_MS       = local.pronounce_timeout_ms
    APP_ENV                    = local.environment
    LOG_FORMAT                 = local.logging.format
  }
  deploy_commit = data.external.origin_head.result.sha
}

module "web_app" {
  source         = "../../modules/application"
  name           = local.apps.web.name
  project_uuid   = coolify_project.engflex.uuid
  server_uuid    = var.server_uuid
  git_repository = var.git_repository
  git_branch     = var.git_branch

  base_directory          = "/"
  build_pack              = "dockercompose"
  docker_compose_location = "/infrastructure/composes/web/compose.yaml"
  ports_exposes           = local.apps.web.port

  docker_compose_domains = jsonencode([{ name = "web", domain = join(",", concat([local.web_url], var.cloudflare_enabled ? ["https://${module.tunnel[0].hostname}"] : [])) }])
  health_check_path      = "/"

  # Tunnel (web-only front door): the tunnel delivers plain HTTP to the
  # proxy, so the https redirect stays off — otherwise origin 301s loop
  # into TOO_MANY_REDIRECTS (Coolify guide step 5). Local routes are
  # plain HTTP by intent (tunnel-only + localhost, no local TLS).
  is_force_https_enabled = !var.cloudflare_enabled
  environment_variables = {
    API_INTERNAL_URL           = "http://${local.apps.api.name}:${local.apps.api.port}/api/v1"
    MEDIA_INTERNAL_ENDPOINT    = "http://${local.minio_net.host}:${local.minio_net.port}"
    MEDIA_BUCKET               = local.minio_net.bucket
    VITE_CLERK_PUBLISHABLE_KEY = var.clerk_publishable_key
    VITE_SENTRY_DSN            = module.sentry.dsns["web"]
    CLERK_SECRET_KEY           = var.clerk_secret_key
  }
  deploy_commit = data.external.origin_head.result.sha
}

module "voice_app" {
  source                  = "../../modules/application"
  name                    = local.apps.voice.name
  project_uuid            = coolify_project.engflex.uuid
  server_uuid             = var.server_uuid
  git_repository          = var.git_repository
  git_branch              = var.git_branch
  build_pack              = "dockercompose"
  base_directory          = "/"
  docker_compose_location = "/infrastructure/composes/voice/compose.yaml"
  ports_exposes           = local.apps.voice.port
  domains                 = []
  health_check_path       = "/status"
  # No Traefik domains (internal-only) — and no TLS anywhere local,
  # so the redirect stays off like api/web (module default is true).
  is_force_https_enabled = false
  # Bulk-managed env: values go in via PATCH bulk,
  # rollout via the module-owned deployment (depends_on orders rows first).
  environment_variables = {
    API_URL          = "http://${local.apps.api.name}:${local.apps.api.port}"
    INTERNAL_SECRET  = var.voice_internal_secret
    STT_API_KEY      = var.stt_api_key
    LLM_API_KEY      = var.llm_api_key
    LLM_BASE_URL     = var.llm_base_url
    LLM_MODEL        = var.llm_model
    SENTRY_DSN       = module.sentry.dsns["voice"]
    TTS_API_KEY      = var.tts_api_key
    VOICE_MODEL      = var.voice_model
    MALE_VOICE_ID    = var.male_voice_id
    FEMALE_VOICE_ID  = var.female_voice_id
    MAX_DURATION_SEC = local.voice_limits.max_duration_sec
    IDLE_TIMEOUT_SEC = local.voice_limits.idle_timeout_sec
    PIPECAT_ICE_SERVERS = jsonencode([
      { urls = [local.stun.url] },
      { urls = ["turn:free.expressturn.com:3478"], username = var.expressturn_username, credential = var.expressturn_credential },
    ])
  }
  deploy_commit      = data.external.origin_head.result.sha
  deployment_timeout = "30m"
}

module "pronounce_app" {
  source              = "../../modules/application"
  name                = local.apps.pronounce.name
  project_uuid        = coolify_project.engflex.uuid
  server_uuid         = var.server_uuid
  git_repository      = var.git_repository
  git_branch          = var.git_branch
  base_directory      = "/"
  dockerfile_location = "/infrastructure/composes/pronounce/Dockerfile"
  ports_exposes       = local.apps.pronounce.port
  domains             = []
  health_check_path   = "/health"
  # Domainless internal app — redirect off like the other three
  # (module default is true; nothing serves TLS locally).
  is_force_https_enabled = false
  # Stable internal DNS on the shared coolify network (same rule as api).
  network_aliases       = "engflex-pronounce"
  environment_variables = {}
  deploy_commit         = data.external.origin_head.result.sha
}

module "minio" {
  source        = "../../modules/minio"
  name          = local.services.minio.name
  project_uuid  = coolify_project.engflex.uuid
  server_uuid   = var.server_uuid
  minio_host    = local.minio_host
  root_password = var.minio_root_password
  bucket        = local.minio_net.bucket
  environment_variables = {
    SERVICE_USER_MINIO         = var.minio_root_user
    MINIO_SERVER_URL           = "http://${local.minio_host}"
    MINIO_BROWSER_REDIRECT_URL = "http://${local.minio_host}"
    SERVICE_PASSWORD_MINIO     = var.minio_root_password
  }
}

# Cloudflare Tunnel in front of the Coolify proxy (all-resource pattern).
# Disabled by default — set cloudflare_enabled = true plus zone vars once
# DNS is on Cloudflare. First boot is TF-owned: the module's resource_action
# fires `start` after the token env exists (instant_deploy = false).
module "tunnel" {
  count                  = var.cloudflare_enabled ? 1 : 0
  source                 = "../../modules/tunnel"
  name                   = local.services.tunnel.name
  project_uuid           = coolify_project.engflex.uuid
  server_uuid            = var.server_uuid
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
  environment  = local.environment

  projects = local.sentry_projects
}
