resource "cloudflare_zero_trust_tunnel_cloudflared" "this" {
  account_id = var.cloudflare_account_id
  name       = var.cloudflare_tunnel_name
  config_src = "cloudflare"
}

# Token the agent runs with. Read, never stored anywhere except the
# service env below and TF state.
data "cloudflare_zero_trust_tunnel_cloudflared_token" "this" {
  account_id = var.cloudflare_account_id
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.this.id
}

resource "cloudflare_dns_record" "this" {
  zone_id = var.cloudflare_zone_id
  name    = var.app_name
  content = "${cloudflare_zero_trust_tunnel_cloudflared.this.id}.cfargotunnel.com"
  type    = "CNAME"
  ttl     = 1
  proxied = true
}

# Published application route: everything for this hostname goes to the
# Coolify proxy.
resource "cloudflare_zero_trust_tunnel_cloudflared_config" "this" {
  tunnel_id  = cloudflare_zero_trust_tunnel_cloudflared.this.id
  account_id = var.cloudflare_account_id
  config = {
    ingress = [
      {
        hostname = "${var.app_name}.${var.cloudflare_zone}"
        service  = "http://localhost:80"
      },
      {
        service = "http_status:404"
      }
    ]
  }
}

# Access gate in front of the hostname (drop these two resources if the
# app should be public like the Coolify walkthrough).
resource "cloudflare_zero_trust_access_policy" "allow_emails" {
  account_id = var.cloudflare_account_id
  name       = "Allow email addresses"
  decision   = "allow"
  include = [
    {
      email = {
        email = var.cloudflare_email
      }
    }
  ]
}

# resource "cloudflare_zero_trust_access_application" "this" {
#   account_id = var.cloudflare_account_id
#   type       = "self_hosted"
#   name       = "Access application for ${var.app_name}.${var.cloudflare_zone}"
#   domain     = "${var.app_name}.${var.cloudflare_zone}"
#   policies = [
#     {
#       id         = cloudflare_zero_trust_access_policy.allow_emails.id
#       precedence = 1
#     }
#   ]
# }

resource "coolify_service" "this" {
  name               = var.name
  project_uuid       = var.project_uuid
  server_uuid        = var.server_uuid
  docker_compose_raw = <<-EOT
    services:
      cloudflared:
        container_name: cloudflare-tunnel
        image: 127.0.0.1:5000/cloudflared:latest
        restart: unless-stopped
        network_mode: host
        command: tunnel --no-autoupdate run
        environment:
          # Dollar-ref reads the LOCAL row (bulk-filled). Project-refs read
          # ONLY project-level vars — the two ref forms never mix (mixing
          # yields a fallback token instead of the real one). No TF env
          # resource for this key (birth placeholder 409s POSTs); values go
          # in via service-bulk.
          - 'TUNNEL_TOKEN=$${CLOUDFLARE_TUNNEL_TOKEN}'
        healthcheck:
          test: ["CMD", "cloudflared", "--version"]
          interval: 5s
          timeout: 20s
          retries: 10
  EOT
  # Rollout owner is the action below; instant stays false so a fresh apply
  # boots the agent exactly once, after the token env exists (never races it).
  instant_deploy = var.instant_deploy
}

# Bulk upsert owns the agent token (same rule as apps: no per-key
# resources). The value reads the data source DIRECTLY, not the module
# output — same bytes, no self-dependency cycle.
resource "coolify_envs_bulk" "this" {
  resource_type = "service"
  resource_uuid = coolify_service.this.uuid
  variables = {
    CLOUDFLARE_TUNNEL_TOKEN = data.cloudflare_zero_trust_tunnel_cloudflared_token.this.token
  }
}

# Rollout owner. `start` fires on birth — after the bulk above, so the token
# exists — and re-fires on token change. `start` materializes
# never-deployed services (it booted the first agent from nothing).
resource "coolify_resource_action" "boot" {
  action        = "start"
  resource_type = "service"
  resource_uuid = coolify_service.this.uuid
  triggers      = { token = sha256(data.cloudflare_zero_trust_tunnel_cloudflared_token.this.token) }
  depends_on    = [coolify_service.this, coolify_envs_bulk.this]
}
