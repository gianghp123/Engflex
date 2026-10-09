# urls entry is keyed by the catalog compose container name ("minio"), not
# the service resource name.
resource "coolify_service" "this" {
  name           = var.name
  project_uuid   = var.project_uuid
  server_uuid    = var.server_uuid
  type           = "minio"
  urls           = [{ name = "minio", url = "https://${var.minio_host}" }]
  instant_deploy = true
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
