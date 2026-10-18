variable "name" {
  type = string
}

variable "project_uuid" {
  type = string
}

variable "server_uuid" {
  type = string
}

variable "git_repository" {
  type = string
  # Overridden per environment (see environments/*/terraform.tfvars).
  default = "http://engflex-git.192-168-1-10.sslip.io:9419/engflex.git"
}

variable "git_branch" {
  description = "Git branch Coolify checks out (set per environment)"
  type        = string
}

variable "build_pack" {
  type    = string
  default = "dockerfile"
}

variable "base_directory" {
  type    = string
  default = "/"
}

variable "dockerfile_location" {
  type    = string
  default = null
}

variable "docker_compose_location" {
  type    = string
  default = null
}

variable "docker_compose_domains" {
  description = "Compose-service domain map as JSON (Coolify accepts it only after the first deploy populates docker_compose_raw); null omits it"
  type        = string
  default     = null
}

variable "ports_exposes" {
  type = string
}

variable "domains" {
  type    = list(string)
  default = []
}

# call sites use the default.
variable "is_force_https_enabled" {
  type    = bool
  default = true
}

variable "health_check_path" {
  type = string
}

# Mirrors Coolify's default start period; raise it for apps that refuse
# traffic until warm (the pronounce model boot) so a deployment is not marked
# unhealthy while the app is still loading.
variable "health_check_start_period" {
  type    = number
  default = 5
}

# Stable DNS name for this app on the shared `coolify` network
# (explicit alias: bare resource names do not resolve there).
variable "network_aliases" {
  type    = string
  default = null
}

variable "custom_labels" {
  description = "Extra Docker labels for the container (e.g. Traefik routers/middlewares); newline-separated key=value lines, omitted when empty"
  type        = string
  default     = ""
}

variable "environment_variables" {
  type      = map(string)
  default   = {}
  sensitive = true
}

variable "instant_deploy" {
  description = "Boot immediately on creation (default false: TF-owned deployments roll out)"
  type        = bool
  default     = false
}

variable "deploy_commit" {
  description = "Origin HEAD of the branch Coolify checks out; a move redeploys (code rollout)"
  type        = string
}

variable "deployment_timeout" {
  description = "wait_for_completion ceiling per app (cold voice builds exceeded 10m)"
  type        = string
  default     = "15m"
}
