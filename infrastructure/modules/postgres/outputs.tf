output "uuid" {
  value = coolify_database_postgresql.this.uuid
}

# Same-project containers reach this DB at its resource name; 5432 is the
# container port (public_port is the host mapping for goose only).
output "internal_url" {
  value     = "postgresql://${var.db_user}:${urlencode(var.db_password)}@${var.name}:5432/${var.db_name}?sslmode=disable"
  sensitive = true
}

output "public_url" {
  value     = "postgresql://${var.db_user}:${urlencode(var.db_password)}@localhost:${var.public_port}/${var.db_name}?sslmode=disable"
  sensitive = true
}
