locals {
  ip_dashed  = replace(var.server_ip, ".", "-")
  api_host   = "engflex-api.${local.ip_dashed}.sslip.io"
  web_host   = "engflex-web.${local.ip_dashed}.sslip.io"
  minio_host = "engflex-minio.${local.ip_dashed}.sslip.io"

  # Flipped to https in Task 14; internal container traffic stays http by intent.
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
  source       = "../../modules/application"
  name         = "engflex-api"
  project_uuid = coolify_project.engflex.uuid
  server_uuid  = var.server_uuid

  # Repo-root context: dockerfile_location is repo-relative.
  base_directory      = "/"
  dockerfile_location = "/apps/api/Dockerfile"
  ports_exposes       = "8000"

  domains           = [local.api_url]
  health_check_path = "/health"

  environment_variables = {
    DATABASE_URL        = module.postgres.internal_url
    REDIS_URL           = module.redis.internal_url
    REDIS_KEY_PREFIX    = "engflex"
    MINIO_ENDPOINT      = "http://minio:9000"
    MINIO_ROOT_USER     = var.minio_root_user
    MINIO_ROOT_PASSWORD = var.minio_root_password
    MINIO_BUCKET        = "lesson-audio"
    MEDIA_BASE_URL      = local.minio_url
    CLERK_SECRET_KEY    = var.clerk_secret_key
    CORS_ORIGINS        = local.web_url
    # Trust the Coolify proxy network for X-Forwarded-For so the IP-keyed
    # limiter buckets by real client IP (both v4 and v6 subnets).
    RATE_LIMIT_TRUSTED_PROXIES = "10.0.1.0/24,fdf6:8678:91f4::/64"
    LLM_API_KEY                = var.llm_api_key
    LLM_BASE_URL               = var.llm_base_url
    LLM_MODEL                  = var.llm_model
    VOICE_SERVICE_URL          = "http://engflex-voice:7860"
    VOICE_INTERNAL_SECRET      = var.voice_internal_secret
    VOICE_MAX_DURATION_SEC     = "300"
    PRONOUNCE_SERVICE_URL      = "http://engflex-pronounce:8000"
    PRONOUNCE_TIMEOUT_MS       = "120000"
    APP_ENV                    = "production"
    LOG_FORMAT                 = "json"
  }
}

module "web_app" {
  source       = "../../modules/application"
  name         = "engflex-web"
  project_uuid = coolify_project.engflex.uuid
  server_uuid  = var.server_uuid

  base_directory      = "/"
  dockerfile_location = "/apps/web/Dockerfile"
  ports_exposes       = "3000"

  domains           = [local.web_url]
  health_check_path = "/"

  environment_variables = {
    VITE_API_URL               = "${local.api_url}/api/v1"
    VITE_CLERK_PUBLISHABLE_KEY = var.clerk_publishable_key
    CLERK_SECRET_KEY           = var.clerk_secret_key
  }
}

module "voice_app" {
  source                  = "../../modules/application"
  name                    = "engflex-voice"
  project_uuid            = coolify_project.engflex.uuid
  server_uuid             = var.server_uuid
  build_pack              = "dockercompose"
  base_directory          = "/"
  docker_compose_location = "/infrastructure/voice/compose.yaml"
  ports_exposes           = "7860"
  domains                 = []
  health_check_path       = "/status"
  environment_variables = {
    API_URL          = "http://engflex-api:8000"
    INTERNAL_SECRET  = var.voice_internal_secret
    STT_API_KEY      = var.stt_api_key
    LLM_API_KEY      = var.llm_api_key
    LLM_BASE_URL     = var.llm_base_url
    LLM_MODEL        = var.llm_model
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
}

module "pronounce_app" {
  source                = "../../modules/application"
  name                  = "engflex-pronounce"
  project_uuid          = coolify_project.engflex.uuid
  server_uuid           = var.server_uuid
  base_directory        = "/"
  dockerfile_location   = "/infrastructure/pronounce/Dockerfile"
  ports_exposes         = "8000"
  domains               = []
  health_check_path     = "/health"
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
