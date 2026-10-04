resource "volcano_database_branch" "feature" {
  project_id    = volcano_project.example.id
  database_name = volcano_database.main.name
  name          = "feature-x"
  ttl_seconds   = 3600
}
