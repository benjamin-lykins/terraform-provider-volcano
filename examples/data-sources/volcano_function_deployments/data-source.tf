data "volcano_function_deployments" "api" {
  project_id  = volcano_project.example.id
  function_id = volcano_function.api.id
}
