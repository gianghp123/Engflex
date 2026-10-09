output "coolify_endpoint" {
  value = var.coolify_endpoint
}

output "coolify_token" {
  value     = var.coolify_token
  sensitive = true
}

output "server_ip_dashed" {
  value = local.ip_dashed
}

output "project_uuid" {
  value = coolify_project.engflex.uuid
}

output "postgres_uuid" {
  value = module.postgres.uuid
}

output "postgres_public_url" {
  value     = module.postgres.public_url
  sensitive = true
}

output "postgres_internal_url" {
  value     = module.postgres.internal_url
  sensitive = true
}

output "redis_uuid" {
  value = module.redis.uuid
}

output "redis_internal_url" {
  value     = module.redis.internal_url
  sensitive = true
}

output "api_app_uuid" {
  value = module.api_app.uuid
}

output "web_app_uuid" {
  value = module.web_app.uuid
}

output "voice_app_uuid" {
  value = module.voice_app.uuid
}

output "pronounce_app_uuid" {
  value = module.pronounce_app.uuid
}

output "api_host" {
  value = local.api_host
}

output "web_host" {
  value = local.web_host
}

output "api_health_url" {
  value = "${local.api_url}/health"
}

output "minio_uuid" {
  value = module.minio.uuid
}

output "minio_host" {
  value = local.minio_host
}

output "minio_root_user" {
  value = var.minio_root_user
}

output "minio_root_password" {
  value     = var.minio_root_password
  sensitive = true
}
