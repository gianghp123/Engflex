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
  docker_compose_domains  = var.docker_compose_domains
  ports_exposes           = var.ports_exposes
  custom_network_aliases  = var.network_aliases
  custom_labels           = var.custom_labels != "" ? var.custom_labels : null

  domains                = length(var.domains) > 0 ? join(",", var.domains) : null
  autogenerate_domain    = false
  is_force_https_enabled = var.is_force_https_enabled

  health_check_enabled = true
  health_check_path    = var.health_check_path
  health_check_scheme  = "http"
  # Apps with a slow warm boot (model loading behind a gated /health) raise
  # this; the default equals the value Coolify already applies, so passing it
  # explicitly is a no-op for everything else.
  health_check_start_period = var.health_check_start_period

  # Rollout is TF-owned (coolify_deployment.this below): instant stays false.
  # instant=true + any trigger = duplicate deploys (two rollout owners race).
  instant_deploy = var.instant_deploy
}

# Bulk upsert owns values: POST-per-key 409s on birth placeholders, so no
# per-key coolify_environment_variable exists here.
# Count-gated: the API 400s on empty data, so apps without env (pronounce)
# get zero bulk resources instead of an empty PATCH.
resource "coolify_envs_bulk" "this" {
  count         = length(var.environment_variables) > 0 ? 1 : 0
  resource_type = "application"
  resource_uuid = coolify_application.this.uuid
  variables     = var.environment_variables
}

# Declarative rollout: replaced (exactly one deployment) when the app, its
# env, or the origin commit moves; anything else = no-op. depends_on orders
# bulk rows before the deploy (kills the instant-vs-bulk race structurally);
# replace_triggered_by is same-module-legal for both (cross-module refs are
# rejected by core). Services roll out via coolify_resource_action in their
# own modules.
resource "coolify_deployment" "this" {
  application_uuid = coolify_application.this.uuid
  triggers = {
    commit = var.deploy_commit
    env    = sha256(jsonencode(var.environment_variables))
    labels = sha256(var.custom_labels)
  }
  depends_on          = [coolify_envs_bulk.this]
  wait_for_completion = true
  timeouts            = { create = var.deployment_timeout }

  lifecycle {
    # App-structural changes only: replace_triggered_by cannot reference a
    # zero-count resource, so env-driven replacement rides triggers.env
    # (same map the bulk writes — always in sync). Manual out-of-TF row edits are the only uncovered vector
    # (accepted: bulk is TF-owned, nothing else writes these rows).
    replace_triggered_by = [coolify_application.this]
  }
}
