resource "sifflet_user" "example" {
  email = "user@example.com"
  name  = "Example User"
  role  = "EDITOR"
  # No matter the system role, all users must have at least one domain role.
  permissions = [{
    domain_id   = "4b0968b9-3a39-46fc-9480-cd117d8a0fbe"
    domain_role = "VIEWER"
  }]
}
