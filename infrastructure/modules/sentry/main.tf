# Sentry projects + client keys for the instrumented services, provisioned
# per deployment environment. One team owns every project the stack reports
# into. A future staging environment reuses this module with its own
# environment value — the suffix exists because renaming a project's slug
# replaces its client key (see the provider's sentry_key docs). The team
# itself must NOT be re-created there: sentry_team names are unique per org,
# so a second environment must either import the existing team or use a
# distinct team name.
resource "sentry_team" "this" {
  organization = var.organization
  name         = var.team
}

# Removing an entry from var.projects destroys that project and its issue
# history on the next apply — deliberate Terraform semantics; treat map
# removals with care.
resource "sentry_project" "this" {
  for_each = var.projects

  organization = var.organization
  teams        = [sentry_team.this.slug]
  name         = "${each.value.name}-${var.environment}"
  slug         = "${each.value.slug}-${var.environment}"
  platform     = each.value.platform
}

resource "sentry_key" "this" {
  for_each = var.projects

  organization = var.organization
  project      = sentry_project.this[each.key].slug
  name         = "${sentry_project.this[each.key].name} Client Key"
}
