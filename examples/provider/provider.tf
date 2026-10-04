terraform {
  required_providers {
    volcano = {
      source = "benjamin-lykins/volcano"
    }
  }
}

# token may also be supplied via the VOLCANO_TOKEN environment variable,
# and endpoint via VOLCANO_API_URL.
provider "volcano" {
}
