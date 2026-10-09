variable "name" {
  type = string
}

variable "project_uuid" {
  type = string
}

variable "server_uuid" {
  type = string
}

variable "redis_password" {
  type      = string
  sensitive = true
}
