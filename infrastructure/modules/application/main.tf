resource "coolify_application" "this" {
  name         = var.name
  project_uuid = var.project_uuid
  server_uuid  = var.server_uuid

  git_repository = var.git_repository
  git_branch     = var.git_branch

  # The local dumb-HTTP git mirror cannot shallow-clone (git limitation).
  is_git_shallow_clone_enabled = false

  build_pack              = var.build_pack
  base_directory          = var.base_directory
  dockerfile_location     = var.build_pack == "dockerfile" ? var.dockerfile_location : null
  docker_compose_location = var.build_pack == "dockercompose" ? var.docker_compose_location : null
  ports_exposes           = var.ports_exposes
  custom_network_aliases  = var.network_aliases

  domains                = length(var.domains) > 0 ? join(",", var.domains) : null
  autogenerate_domain    = false
  is_force_https_enabled = var.is_force_https_enabled

  health_check_enabled = true
  health_check_path    = var.health_check_path
  health_check_scheme  = "http"

  # Deploys trigger via make deploy (deploy API), not on config apply.
  instant_deploy = false
}

# for_each runs over nonsensitive() because the map is sensitive; values are
# still read from the sensitive map so secrecy is preserved. Both flags are
# true so VITE_* bake into the image build and secrets reach runtime.
resource "coolify_environment_variable" "this" {
  for_each         = nonsensitive(var.environment_variables)
  application_uuid = coolify_application.this.uuid
  key              = each.key
  value            = var.environment_variables[each.key]
  is_build         = true
  is_runtime       = true
}
