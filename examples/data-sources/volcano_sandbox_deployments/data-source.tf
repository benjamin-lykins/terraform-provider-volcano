data "volcano_sandbox_deployments" "dev" {
  project_id = volcano_project.example.id
  sandbox_id = volcano_sandbox.dev.id
}
