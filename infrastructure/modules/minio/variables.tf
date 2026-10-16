variable "name" {
  type = string
}

variable "project_uuid" {
  type = string
}

variable "server_uuid" {
  type = string
}

# Public hostname WITHOUT scheme, e.g. engflex-minio.192-168-1-50.sslip.io.
variable "minio_host" {
  type = string
}

variable "root_password" {
  type      = string
  sensitive = true
}

variable "environment_variables" {
  description = "Service env map, applied via bulk upsert (no per-key resources: birth placeholders 409 POSTs)"
  type        = map(string)
  default     = {}
}
