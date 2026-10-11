# Bulk-managed env (2026-10-10): TF owns values via PATCH-bulk upsert —
# birth placeholders never 409 (POST-only was the whole saga).
# NOTE: jsonencode sorts object keys; the live PIPECAT row was written by
# Python (insertion order) so the FIRST apply shows a 1-key diff that only
# reorders JSON (functionally identical), then plans go clean.
locals {
  ip_dashed_bulk  = replace(var.server_ip, ".", "-")
  api_url_bulk    = "https://engflex-api.${local.ip_dashed_bulk}.sslip.io"
  web_url_bulk    = "https://engflex-web.${local.ip_dashed_bulk}.sslip.io"
  minio_host_bulk = "engflex-minio.${local.ip_dashed_bulk}.sslip.io"

  env_api = {
    DATABASE_URL               = module.postgres.internal_url
    DB_HOST                    = module.postgres.uuid
    DB_PORT                    = "5432"
    DB_USER                    = "engflex"
    DB_PASSWORD                = var.db_password
    DB_NAME                    = "engflex"
    DB_SSLMODE                 = "disable"
    REDIS_URL                  = module.redis.internal_url
    REDIS_KEY_PREFIX           = "engflex"
    MINIO_ENDPOINT             = "http://minio:9000"
    MINIO_ROOT_USER            = "engflex"
    MINIO_ROOT_PASSWORD        = var.minio_root_password
    MINIO_BUCKET               = "lesson-audio"
    MEDIA_BASE_URL             = "https://${local.minio_host_bulk}"
    CLERK_SECRET_KEY           = var.clerk_secret_key
    CORS_ORIGINS               = join(",", [local.web_url_bulk, local.api_url_bulk])
    RATE_LIMIT_TRUSTED_PROXIES = "10.0.1.0/24,fdf6:8678:91f4::/64"
    LLM_API_KEY                = var.llm_api_key
    LLM_BASE_URL               = var.llm_base_url
    LLM_MODEL                  = var.llm_model
    SENTRY_DSN                 = module.sentry.dsns["api"]
    VOICE_SERVICE_URL          = "http://engflex-voice:7860"
    VOICE_INTERNAL_SECRET      = var.voice_internal_secret
    VOICE_MAX_DURATION_SEC     = "300"
    PRONOUNCE_SERVICE_URL      = "http://engflex-pronounce:8000"
    PRONOUNCE_TIMEOUT_MS       = "120000"
    APP_ENV                    = "production"
    LOG_FORMAT                 = "json"
  }

  env_web = {
    API_INTERNAL_URL           = "http://engflex-api:8000/api/v1"
    VITE_CLERK_PUBLISHABLE_KEY = var.clerk_publishable_key
    VITE_SENTRY_DSN            = module.sentry.dsns["web"]
    CLERK_SECRET_KEY           = var.clerk_secret_key
  }

  env_voice = {
    API_URL          = "http://engflex-api:8000"
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
    MAX_DURATION_SEC = "300"
    IDLE_TIMEOUT_SEC = "300"
    PIPECAT_ICE_SERVERS = jsonencode([
      { urls = ["stun:stun.l.google.com:19302"] },
      { urls = ["turn:${var.server_ip}:3478"], username = "engflex", credential = var.turn_secret },
    ])
    SERVER_IP   = var.server_ip
    TURN_SECRET = var.turn_secret
  }

  env_minio = {
    SERVICE_USER_MINIO         = "engflex"
    MINIO_SERVER_URL           = "https://${local.minio_host_bulk}"
    MINIO_BROWSER_REDIRECT_URL = "https://${local.minio_host_bulk}"
    SERVICE_PASSWORD_MINIO     = var.minio_root_password
  }
}

resource "coolify_envs_bulk" "api" {
  resource_type = "application"
  resource_uuid = module.api_app.uuid
  variables     = local.env_api
}

resource "coolify_envs_bulk" "web" {
  resource_type = "application"
  resource_uuid = module.web_app.uuid
  variables     = local.env_web
}

resource "coolify_envs_bulk" "voice" {
  resource_type = "application"
  resource_uuid = module.voice_app.uuid
  variables     = local.env_voice
}

resource "coolify_envs_bulk" "minio" {
  resource_type = "service"
  resource_uuid = module.minio.uuid
  variables     = local.env_minio
}

resource "coolify_envs_bulk" "tunnel" {
  count         = var.cloudflare_enabled ? 1 : 0
  resource_type = "service"
  resource_uuid = module.tunnel[0].uuid
  variables = {
    CLOUDFLARE_TUNNEL_TOKEN = module.tunnel[0].tunnel_token
  }
}
