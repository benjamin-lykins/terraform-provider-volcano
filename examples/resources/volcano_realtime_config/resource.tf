resource "volcano_realtime_config" "example" {
  project_id               = volcano_project.example.id
  enabled                  = true
  broadcast_enabled        = true
  presence_enabled         = true
  postgres_changes_enabled = false
}
