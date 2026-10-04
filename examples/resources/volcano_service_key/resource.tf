resource "volcano_service_key" "jobs" {
  project_id  = volcano_project.example.id
  name        = "background-jobs"
  permissions = ["functions.invoke", "locks.manage"]
}
