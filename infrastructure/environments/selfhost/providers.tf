terraform {
  required_version = ">= 1.6"

  required_providers {
    coolify = {
      source  = "coolify-terraform/coolify"
      version = ">= 0.1"
    }
    sentry = {
      source  = "jianyuan/sentry"
      version = "~> 0.15.0"
    }
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = ">= 5.8.2"
    }
  }

  backend "local" {}
}

provider "coolify" {
  endpoint = var.coolify_endpoint
  token    = var.coolify_token
}

# Provider auth: an org-scoped internal-integration token (README Ops notes).
# Stored only in the gitignored terraform.tfvars.
provider "sentry" {
  token = var.sentry_auth_token
}

provider "cloudflare" {
  api_token = var.cloudflare_token
}
