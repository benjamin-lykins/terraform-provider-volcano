resource "volcano_function_scheduler" "every_5_min" {
  project_id      = volcano_project.example.id
  function_id     = volcano_function.api.id
  name            = "every-5-min"
  cron_expression = "*/5 * * * *"
  payload         = jsonencode({ source = "scheduler" })
}
