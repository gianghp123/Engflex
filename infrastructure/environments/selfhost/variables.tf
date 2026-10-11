variable "coolify_endpoint" {
  type    = string
  default = "http://localhost:8000"
}

variable "coolify_token" {
  type      = string
  sensitive = true
}

variable "server_uuid" {
  type = string
}

variable "git_repository" {
  description = "Git source Coolify clones (per environment; retired the local mirror 2026-10-10)"
  type        = string
}

variable "server_ip" {
  type = string # e.g. "192.168.1.50"
}

variable "db_public_port" {
  type    = number
  default = 5433
}

variable "db_name" {
  type    = string
  default = "engflex"
}

variable "db_user" {
  type    = string
  default = "engflex"
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "redis_password" {
  type      = string
  sensitive = true
}

variable "minio_root_user" {
  type    = string
  default = "engflex"
}

variable "minio_root_password" {
  type      = string
  sensitive = true
}

variable "clerk_secret_key" {
  type      = string
  sensitive = true
}

variable "clerk_publishable_key" {
  type = string
}

variable "llm_api_key" {
  type      = string
  sensitive = true
}

variable "llm_base_url" {
  type = string
}

variable "llm_model" {
  type = string
}

variable "voice_internal_secret" {
  type      = string
  sensitive = true
}

variable "stt_api_key" {
  type      = string
  sensitive = true
}

variable "tts_api_key" {
  type      = string
  sensitive = true
}

variable "voice_model" {
  type = string
}

variable "male_voice_id" {
  type = string
}

variable "female_voice_id" {
  type = string
}

variable "turn_secret" {
  type      = string
  sensitive = true
}

# Extra origins appended to the API's CORS_ORIGINS list (which always
# includes the web frontend and the API origin itself for Swagger UI).
variable "cors_extra_origins" {
  type    = list(string)
  default = []
}

# Sentry provisioning (jianyuan/sentry provider): the stack creates its own
# projects and client keys. The auth token is an org-scoped
# internal-integration token kept in the gitignored terraform.tfvars; the
# organization must exist, the team is created by the module.
variable "sentry_organization" {
  type = string
}

variable "sentry_team" {
  type = string
}

variable "sentry_auth_token" {
  type      = string
  sensitive = true
}

# Cloudflare variables
variable "cloudflare_account_id" {
  description = "Account ID for your Cloudflare account"
  type        = string
  sensitive   = true
}

variable "cloudflare_email" {
  description = "Email address for your Cloudflare account"
  type        = string
  sensitive   = true
}

variable "cloudflare_token" {
  description = "Cloudflare API token"
  type        = string
  sensitive   = true
}

variable "cloudflare_enabled" {
  description = "Create the Cloudflare Tunnel + cloudflared service (needs zone vars below)"
  type        = bool
  default     = false
}

variable "cloudflare_tunnel_name" {
  description = "Display name for the cloudflared tunnel"
  type        = string
  default     = "engflex selfhost"
}

variable "cloudflare_zone" {
  description = "Cloudflare zone name (root domain, DNS must be on Cloudflare)"
  type        = string
  default     = ""
}

variable "cloudflare_zone_id" {
  description = "Zone ID for cloudflare_zone"
  type        = string
  default     = ""
  sensitive   = true
}

variable "cloudflare_app_name" {
  description = "Subdomain exposed through the tunnel"
  type        = string
  default     = "engflex"
}
