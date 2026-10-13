output "dsns" {
  description = "Public DSN per project, keyed like var.projects. Sensitive: only env-var injection should read it."
  value       = { for k, key in sentry_key.this : k => key.dsn["public"] }
  sensitive   = true
}

output "team_slug" {
  description = "Slug of the created team."
  value       = sentry_team.this.slug
}
