terraform {
  required_version = ">= 1.6"

  required_providers {
    coolify = {
      source  = "coolify-terraform/coolify"
      version = ">= 0.1"
    }
  }

  backend "local" {}
}

provider "coolify" {
  endpoint = var.coolify_endpoint
  token    = var.coolify_token
}
