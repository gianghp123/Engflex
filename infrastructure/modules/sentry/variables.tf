variable "organization" {
  type        = string
  description = "Sentry organization slug (must already exist)."
}

variable "team" {
  type        = string
  description = "Team name to create and attach every project to."
}

variable "environment" {
  type        = string
  description = "Deployment environment; appended to every project name and slug."
}

variable "projects" {
  description = "Projects to create, keyed by short service name (e.g. api, web, voice)."
  type = map(object({
    name     = string
    slug     = string
    platform = string
  }))
}
