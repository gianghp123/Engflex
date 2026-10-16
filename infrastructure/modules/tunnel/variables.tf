variable "name" {
  description = "Coolify service name for the cloudflared agent"
  type        = string
}

variable "project_uuid" {
  type = string
}

variable "server_uuid" {
  type = string
}

variable "cloudflare_tunnel_name" {
  type = string
}

variable "cloudflare_account_id" {
  description = "Account ID for your Cloudflare account"
  type        = string
  sensitive   = true
}

variable "cloudflare_email" {
  description = "Email address allowed through the Access policy"
  type        = string
  sensitive   = true
}

variable "cloudflare_token" {
  description = "Cloudflare API token"
  type        = string
  sensitive   = true
}

variable "cloudflare_zone_id" {
  description = "Zone ID for your Domain"
  type        = string
  sensitive   = true
}

variable "cloudflare_zone" {
  description = "Cloudflare zone name (root domain name)"
  type        = string
}

variable "app_name" {
  description = "Subdomain for this app"
  type        = string
}

variable "instant_deploy" {
  description = "Boot immediately on creation (default false)"
  type        = bool
  default     = false
}
