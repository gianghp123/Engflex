resource "coolify_database_postgresql" "this" {
  name         = var.name
  project_uuid = var.project_uuid
  server_uuid  = var.server_uuid
  image        = "postgres:16"

  postgres_user     = var.db_user
  postgres_password = var.db_password
  postgres_db       = var.db_name

  is_public      = true
  public_port    = var.public_port
  instant_deploy = true
}
