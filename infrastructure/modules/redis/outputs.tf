output "uuid" {
  value = coolify_database_redis.this.uuid
}

output "internal_url" {
  value     = "redis://:${urlencode(var.redis_password)}@${var.name}:6379"
  sensitive = true
}
