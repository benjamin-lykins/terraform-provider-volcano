resource "volcano_durable_function" "workflow" {
  project_id = volcano_project.example.id
  name       = "order-fulfillment"
  runtime    = "python3.13"
  source     = "${path.module}/dist/order-fulfillment.zip"
}
