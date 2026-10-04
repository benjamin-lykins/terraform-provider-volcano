data "volcano_durable_function_deployments" "workflow" {
  project_id  = volcano_project.example.id
  function_id = volcano_durable_function.workflow.id
}
