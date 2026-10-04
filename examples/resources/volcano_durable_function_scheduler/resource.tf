resource "volcano_durable_function_scheduler" "nightly" {
  project_id      = volcano_project.example.id
  function_id     = volcano_durable_function.workflow.id
  name            = "nightly"
  cron_expression = "0 0 * * *"
}
