resource "volcano_database_backup_schedule" "main" {
  project_id    = volcano_project.example.id
  database_name = volcano_database.main.name

  entries = [
    {
      frequency = "daily"
      hour      = 3
    },
    {
      frequency         = "weekly"
      hour              = 4
      day               = 1
      retention_seconds = 30 * 24 * 60 * 60
    },
  ]
}
