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
  # Local dumb-HTTP git mirror on this machine (its `main` = the working
  # snapshot); GitHub push intentionally avoided — see the SDD ledger.
  default = "http://192.168.1.10:9419/engflex.git"
}

variable "git_branch" {
  type    = string
  default = "main"
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

variable "ports_exposes" {
  type = string
}

variable "domains" {
  type    = list(string)
  default = []
}

# Flipped to true in Task 14 (https via Traefik); call sites use the default.
variable "is_force_https_enabled" {
  type    = bool
  default = true
}

variable "health_check_path" {
  type = string
}

variable "environment_variables" {
  type      = map(string)
  default   = {}
  sensitive = true
}
