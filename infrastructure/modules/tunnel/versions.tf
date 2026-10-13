terraform {
  required_providers {
    cloudflare = {
      source  = "cloudflare/cloudflare"
      version = ">= 5.8.2"
    }
    coolify = {
      source  = "coolify-terraform/coolify"
      version = ">= 0.1"
    }
  }
  required_version = ">= 1.2"
}
# NOTE: no provider blocks here by intent — the caller (environments/*)
# configures both providers, so this module stays countable. A local
# provider block would make it a legacy module (count/for_each forbidden).
