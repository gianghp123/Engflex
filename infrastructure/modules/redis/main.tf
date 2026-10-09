resource "coolify_database_redis" "this" {
  name           = var.name
  project_uuid   = var.project_uuid
  server_uuid    = var.server_uuid
  image          = "redis:7-alpine"
  redis_password = var.redis_password
  instant_deploy = true
}
