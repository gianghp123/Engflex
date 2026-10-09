output "uuid" {
  value = coolify_database_postgresql.this.uuid
}

# Coolify 4.4.3 gives standalone-DB containers no name-based DNS (verified:
# `engflex-postgres` is NXDOMAIN on the coolify network; only the container
# uuid resolves). The host part MUST be the database uuid; 5432 is the
# container port (public_port is the host mapping for goose only).
output "internal_url" {
  value     = "postgresql://${var.db_user}:${urlencode(var.db_password)}@${coolify_database_postgresql.this.uuid}:5432/${var.db_name}?sslmode=disable"
  sensitive = true
}

output "public_url" {
  value     = "postgresql://${var.db_user}:${urlencode(var.db_password)}@localhost:${var.public_port}/${var.db_name}?sslmode=disable"
  sensitive = true
}
