output "uuid" {
  description = "Coolify cloudflared service UUID (deploy trigger target)"
  value       = coolify_service.this.uuid
}

output "tunnel_id" {
  description = "ID of the remotely-managed cloudflared tunnel"
  value       = cloudflare_zero_trust_tunnel_cloudflared.this.id
}

output "tunnel_token" {
  description = "Token the agent runs with (consumed by shared.tf, never stored elsewhere)"
  value       = data.cloudflare_zero_trust_tunnel_cloudflared_token.this.token
  sensitive   = true
}

output "hostname" {
  description = "Public hostname routed through the tunnel"
  value       = "${var.app_name}.${var.cloudflare_zone}"
}
