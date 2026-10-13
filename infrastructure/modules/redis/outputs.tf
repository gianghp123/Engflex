output "uuid" {
  value = coolify_database_redis.this.uuid
}

# Same rule as postgres: only the database uuid resolves on the coolify
# network (verified NXDOMAIN for the resource name); 6379 is the container
# port.
output "internal_url" {
  value     = "redis://:${urlencode(var.redis_password)}@${coolify_database_redis.this.uuid}:6379"
  sensitive = true
}
