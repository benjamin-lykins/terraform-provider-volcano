# Terraform Provider for Volcano

A Terraform provider for [Volcano](https://docs.volcano.dev) (api.volcano.dev) — manage projects, databases, functions, durable functions, frontends, storage, auth configuration, sandboxes, and related platform resources as code.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://go.dev/doc/install) >= 1.25 (for building from source — this is `terraform-plugin-framework`'s own minimum)
- A Volcano API token (`VOLCANO_TOKEN`) — a platform token (`pk-`), project access token (`pt-`), or service key, depending on which resources you use. See [Authentication](https://docs.volcano.dev/platform/api-reference/authentication).

## Using the provider

```hcl
terraform {
  required_providers {
    volcano = {
      source = "benjamin-lykins/volcano"
    }
  }
}

provider "volcano" {
  # token and endpoint may also come from VOLCANO_TOKEN / VOLCANO_API_URL
}

resource "volcano_project" "example" {
  name = "my-awesome-app"
}
```

See `examples/` and the generated `docs/` for every resource and data source.

## Building from source

```sh
make build    # go build -o terraform-provider-volcano .
make install  # go install .
```

## Developing

```sh
make lint      # go vet + gofmt check
make test      # unit tests (httptest-mocked, no credentials needed)
make testacc   # acceptance tests against the real API (requires TF_ACC=1 and VOLCANO_TOKEN)
make docs      # regenerate docs/ from templates/ and examples/
```

## Scope

This provider covers every endpoint in Volcano's public API that has a natural create/read/update/delete shape. It intentionally does not model:

- End-user authentication flows (signup/signin/sessions/identities) — these are runtime APIs for the applications you build on Volcano, not provider infrastructure.
- Data-plane operations: database query execution, function invocation, durable-function execution control, sandbox exec/session management.
- Observability/billing data: logs, metrics, usage and stats endpoints.
- One-off actions with no idempotent update semantics: redeploy, password reset, key regeneration.

See the resource index under `docs/` for the full list of what *is* covered.
