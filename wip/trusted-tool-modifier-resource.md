# Trusted Thought Tool Modifier Terraform Resource

Status: WIP provider plan; not implemented.

Repository state verified on 2026-08-20:

- the trusted tool modifier server implementation is merged into Tama's
  `develop` branch;
- the `tama-go` modifier client is implemented on
  `feature/implement-trusted-modifier`, but no released tag contains it yet;
- this provider still depends on `github.com/upmaru/tama-go v0.4.1`; and
- the provider acceptance workflow still runs `ghcr.io/upmaru/tama:0.4-server`,
  which predates the modifier API.

This plan is based on:

- `tama/wip/trusted-tool-modifiers.md`; and
- `tama-go/wip/trusted-tool-modifier-api.md`.

The provider resource must not be described as available until it depends on a
released `tama-go` version containing the client, passes acceptance tests
against a deployed Tama version containing the API, and is itself released.

## Goal

Add a `tama_thought_tool_modifier` resource that provisions deterministic,
trusted argument modifiers on a Tama thought tool.

The initial consumer needs this configuration:

```hcl
resource "tama_thought_tool_modifier" "actor_scope" {
  thought_tool_id   = tama_thought_tool.search_result.id
  index             = 0
  target            = "/body/search/scope/user_id"
  on_missing_parent = "skip"
  on_missing_source = "error"

  source {
    type = "metadata"
    path = "actor_identifier"
  }
}
```

The provider owns Terraform schema, plan validation, CRUD mapping, drift
handling, import, documentation, and acceptance coverage. Tama remains
authoritative for JSON Pointer parsing, callable-schema compatibility, active
index and target conflicts, lifecycle transitions, and runtime execution.

## Naming and ownership

The public resource name is exactly:

```text
tama_thought_tool_modifier
```

Implement it under `tama/tools/modifier`. Do not reuse or rename the existing
`tama_action_modifier` implementation under `tama/motor/modifier`:

- `tama_action_modifier` changes an action's model-visible schema; and
- `tama_thought_tool_modifier` injects trusted runtime metadata into one
  thought tool's effective arguments before request construction.

There is no modifier list endpoint, so this phase adds no data source and no
list operation.

## Upstream API contract

All routes use the authenticated provision API:

| Lifecycle | Client method | Request |
| --- | --- | --- |
| Create or reactivate | `client.Tools.CreateModifier(thoughtToolID, request)` | `POST /provision/tools/:thought_tool_id/modifiers` |
| Read | `client.Tools.GetModifier(id)` | `GET /provision/tools/modifiers/:id` |
| Update | `client.Tools.UpdateModifier(id, request)` | `PATCH /provision/tools/modifiers/:id` |
| Deactivate | `client.Tools.DeleteModifier(id)` | `DELETE /provision/tools/modifiers/:id` |

The API also exposes PUT through `ReplaceModifier`, but Terraform Update should
use PATCH. The immutable `index` must never appear in an update request.

Create can reactivate an exact inactive record and return its existing ID.
Delete is a deactivation, not a physical deletion. Show returns `404` for an
unknown, invalid, or inactive modifier.

A provider credential needs either a broad provision scope or both narrow
capabilities:

```text
provision.all
provision.tools.all

or

provision.tools.modifier.read
provision.tools.modifier.manage
```

The read scope covers refresh and import. The manage scope covers create,
update, and delete.

## Terraform schema

| Attribute | Terraform mode | Validation and lifecycle |
| --- | --- | --- |
| `id` | Computed string | Use state for unknown; populated from Tama. |
| `thought_tool_id` | Required string | Non-empty and `RequiresReplace`; ownership cannot move. |
| `index` | Required integer | Non-negative and `RequiresReplace`; Tama makes it immutable. |
| `target` | Required string | Non-empty and mutable; Tama validates JSON Pointer syntax and schema compatibility. |
| `on_missing_parent` | Required string | Mutable; one of `error` or `skip`. |
| `on_missing_source` | Required string | Mutable; one of `error` or `skip`. |
| `source` | Required single nested block | Mutable; contains required `type` and `path`. |
| `source.type` | Required string | Only `metadata` in version 1. |
| `source.path` | Required string | One of `actor_identifier`, `origin_entity_identifier`, or `current_timestamp`. |
| `provision_state` | Computed string | Expected to be `active` while managed. |

Use the exported `tama-go/tools` constants in validators and request mapping so
the provider does not duplicate wire literals in multiple places.

Follow the repository's existing block-style nested configuration with
`schema.SingleNestedBlock`. Add `objectvalidator.IsRequired()` because the
existing single-block examples are otherwise optional at the framework level.
Model the block as a pointer so absent and configured values are distinguishable
during decoding:

```go
type SourceModel struct {
    Type types.String `tfsdk:"type"`
    Path types.String `tfsdk:"path"`
}

type ResourceModel struct {
    Id              types.String `tfsdk:"id"`
    ThoughtToolId   types.String `tfsdk:"thought_tool_id"`
    Index           types.Int64  `tfsdk:"index"`
    Target          types.String `tfsdk:"target"`
    OnMissingParent types.String `tfsdk:"on_missing_parent"`
    OnMissingSource types.String `tfsdk:"on_missing_source"`
    Source          *SourceModel `tfsdk:"source"`
    ProvisionState  types.String `tfsdk:"provision_state"`
}
```

The client uses Go `int` for `index`, while Terraform uses `int64` and the
provider releases 32-bit binaries. Validate a portable upper bound before
conversion rather than allowing architecture-dependent overflow. Tama stores
the value in a PostgreSQL integer, so `0` through `math.MaxInt32` is the safe
provider range.

Do not implement the target's 512-byte maximum with a character-counting
validator. Either add a focused UTF-8 byte-length validator with matching tests
or, preferably for version 1, leave the maximum and the remaining pointer rules
to Tama. A simple non-empty string validator is sufficient locally.

## Resource implementation

Create `tama/tools/modifier/resource.go` following the adjacent
`tama/tools/input` and `tama/tools/initializer` resource structure.

The implementation should satisfy:

```go
var _ resource.Resource = &Resource{}
var _ resource.ResourceWithImportState = &Resource{}
```

### Configure

Use the standard provider-data type assertion and store `*tama.Client` on the
resource. Do not create a second Tama client or construct HTTP requests in the
provider resource.

### Create

1. Decode the plan into `ResourceModel`.
2. Validate that the required source block is present.
3. Safely convert `index` to `int`.
4. Build `tools.CreateModifierRequest` with every required field.
5. Call `CreateModifier` with `thought_tool_id`.
6. Map the complete returned resource into state, including the nested source
   and the returned ID.

Create must accept exact reactivation. It must not assume a newly generated ID.

### Read and drift handling

Call `GetModifier` using the state ID and map all server fields back into
state. Use `errors.As` against `*tools.Error`:

```go
var apiErr *tools.Error
if errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusNotFound {
    resp.State.RemoveResource(ctx)
    return
}
```

This behavior is required because an externally deactivated modifier is absent
from the active show endpoint. Any non-404 error remains a diagnostic. The
provider currently has no shared typed-404 state-removal helper; keep the first
implementation local to this resource unless a separate refactor is explicitly
scoped.

### Update

Build `tools.UpdateModifierRequest` with exactly these complete mutable values:

- `target`;
- `on_missing_parent`;
- `on_missing_source`; and
- a non-nil `*tools.ModifierSource` containing both source fields.

Do not serialize `index` or `thought_tool_id`. Call `UpdateModifier`, then map
the returned resource into state. A server-side `404` during update is an
operation error rather than silent state removal because the planned mutation
did not occur.

### Delete

Call `DeleteModifier`. A successful response means Terraform can remove the
resource from state; do not write the returned inactive lifecycle state.

Treat a typed `404` as already absent so deletion remains idempotent if an
external deactivation races with destroy. Preserve every other client error as
a diagnostic.

### Import

Import accepts only the modifier ID. Call `GetModifier`, populate every field,
and write the complete model to state. Importing an inactive or unknown ID must
return a not-found diagnostic, not create or reactivate anything.

### State mapping

Use one small helper to map `*tools.Modifier` into `ResourceModel` from Create,
Read, Update, and Import. This prevents nested source or ownership fields from
drifting between lifecycle methods. The helper must always set:

- ID;
- thought-tool ID;
- index;
- target;
- both policies;
- source type and path; and
- provision state.

Logging may include resource IDs, index, target, and allowlisted source names.
The provider never receives the resolved runtime metadata value and must not
add request bodies or credentials to logs.

## Provider and dependency integration

### Upgrade `tama-go`

Provider development can temporarily use:

```go
replace github.com/upmaru/tama-go => ../tama-go
```

The replacement must not appear in the final change. Before merge, update
`go.mod` and `go.sum` to a released `tama-go` version that contains
`tools/modifier.go` and verify the selected tag from repository or registry
state rather than guessing a version.

The current client feature has two compatibility effects that the provider
must address when advancing from v0.4.1:

1. `tama.NewClient` now returns `(*tama.Client, error)`. Update
   `TamaProvider.Configure` to add a bounded provider diagnostic and return when
   client construction fails.
2. The current client module requires Go 1.25. Update the provider `go`
   directive and `.tool-versions` to at least the released client's minimum,
   then let the workflows continue to read the version from `go.mod`.

Run `go mod tidy` only after the released dependency is selected and review all
transitive changes.

### Register the resource

Import the new package in `tama/provider.go` using an unambiguous alias such as
`tool_modifier`, then add `tool_modifier.NewResource` to `Resources` near the
other thought-tool children. Keep the existing motor modifier registered
separately.

Extend `tama/provider_test.go` to assert that both of these metadata names are
registered:

```text
tama_action_modifier
tama_thought_tool_modifier
```

This protects against an import alias or registration mistake and documents
that the resources are distinct.

## Tests

Create `tama/tools/modifier/resource_test.go` and a dedicated OpenAPI fixture
whose callable request body contains the complete nested target
`/body/search/scope/user_id`. Do not rely on an unrelated action schema with
dynamic `additionalProperties`, because Tama intentionally rejects unsupported
target traversal.

### Plan-time validation coverage

Cover diagnostics before any API request for:

- a negative index;
- an index above the portable 32-bit range;
- an empty thought-tool ID or target;
- an invalid missing-parent policy;
- an invalid missing-source policy;
- an absent source block;
- a source type other than `metadata`; and
- a source path outside the three version 1 values.

Verify that changing `thought_tool_id` or `index` plans replacement. Verify that
changing target, either policy, or either source field plans an in-place update.

### Acceptance coverage

Use the repository's `resource.Test` and shared acceptance provider factories.
Cover:

1. create and read with `index = 0` and every state field asserted;
2. import with `ImportStateVerify`;
3. in-place updates to target, policies, and source while preserving the ID;
4. replacement when index changes;
5. delete/deactivation;
6. re-adding the exact configuration after deactivation and proving Tama
   reuses the inactive record's ID;
7. an invalid but non-empty target that reaches Tama and preserves the API
   validation diagnostic;
8. refresh after external deactivation, proving Read removes the resource from
   Terraform state; and
9. an already-inactive modifier during Delete, proving destroy is idempotent.

Conflict and callable-schema compatibility remain server behavior. One invalid
target acceptance case is sufficient to prove the provider preserves the
typed API boundary; do not reimplement Tama's schema traversal or conflict
algorithm in provider tests.

## Acceptance environment gap

`.github/workflows/test.yml` currently runs
`ghcr.io/upmaru/tama:0.4-server`. Replace it with an immutable released tag or
digest that contains:

- the trusted tool modifier migrations;
- the provision routes;
- the read/manage permissions; and
- the merged runtime implementation.

Do not guess the image tag. Verify the release artifact before changing the
workflow. The focused and full acceptance suites cannot be completion evidence
while they run against the old server.

Confirm that the bootstrap credential has `provision.all` or both modifier
read/manage permissions. Keep the existing Terraform version matrix unless the
new nested-block validation demonstrates an actual compatibility failure.

The repository also uses `golangci-lint latest` both locally and in CI. When
the Go toolchain is advanced, pin one verified linter version in
`.tool-versions` and the workflow so local and CI gates cannot silently diverge.

## Documentation

Add:

```text
examples/resources/tama_thought_tool_modifier/resource.tf
docs/resources/thought_tool_modifier.md
```

The example must include a callable schema that actually resolves
`/body/search/scope/user_id`, the thought tool, and the modifier. Explain that:

- `source` is trusted metadata, not a Terraform secret value;
- `on_missing_parent = "skip"` leaves calls without `search.scope` unchanged;
- `on_missing_source = "error"` fails closed when actor metadata is required;
- index and thought-tool ownership require replacement; and
- destroy deactivates the resource while exact recreation can reuse its ID.

Generate the resource page with the existing `make generate` workflow. Do not
hand-edit the schema section under the tfplugindocs marker. Re-run generation
after the final schema and example settle, and confirm a second generation pass
produces no additional diff.

## Implementation sequence

1. Verify a Tama server release/deployment containing the merged modifier API.
2. Verify that the client work is merged and that a separately authorized
   `tama-go` release containing the modifier client exists; record its exact
   version.
3. Advance the provider dependency and Go toolchain, remove any local module
   replacement, and adapt `TamaProvider.Configure` to the error-returning
   constructor.
4. Add the resource model, validators, lifecycle methods, typed 404 handling,
   state mapping, and safe logging under `tama/tools/modifier`.
5. Register the resource and extend provider registration coverage.
6. Add the dedicated callable-schema fixture plus plan and acceptance tests.
7. Update the acceptance server image and verify modifier permissions.
8. Add the Terraform example and regenerate provider documentation.
9. Run focused and full validation, inspect generated and dependency diffs,
   and release the provider only after every gate passes.

## Validation commands

Run commands from the provider repository after loading its environment:

```bash
source .envrc
mise exec -- go mod tidy
mise exec -- make fmt
mise exec -- make lint
mise exec -- make build
mise exec -- make test
mise exec -- go test -race ./...
mise exec -- make generate
mise exec -- git diff --check
```

Run the focused acceptance resource against a disposable compatible Tama
server:

```bash
TF_ACC=1 mise exec -- go test -v ./tama/tools/modifier \
  -run TestAccThoughtToolModifierResource \
  -timeout 120m
```

Then run the full acceptance gate:

```bash
source .envrc
mise exec -- make testacc
```

After the intended generated documentation is present, run `make generate`
again and confirm it is idempotent. CI's generate job is the final drift gate.

## Non-goals

This provider phase does not:

- implement modifier execution or metadata resolution;
- inspect or project callable JSON Schemas locally;
- add a modifier list endpoint or data source;
- expose arbitrary source types or paths;
- change `tama_action_modifier`;
- provision a modifier into a Memovee graph; or
- release against an unreleased client or an old server image.

## Acceptance criteria

The provider implementation is complete only when:

- `tama_thought_tool_modifier` is registered independently from
  `tama_action_modifier`;
- the schema enforces immutable ownership/index, closed policy/source values,
  a required source block, and safe index conversion;
- Create, Read, Update, Delete, and Import use the released `tama-go` Tools
  methods without constructing provider-local HTTP requests;
- Update cannot send index and Create preserves zero as a valid index;
- Read removes externally deactivated modifiers from state on a typed `404`;
- Delete is idempotent for an already absent modifier;
- exact inactive configuration reactivation is covered and preserves the
  server-returned ID;
- the provider handles the released client's `NewClient` error contract and
  satisfies its minimum Go version;
- the acceptance workflow runs a verified Tama image containing the API;
- plan validation, lifecycle, import, replacement, API-error, drift, and
  reactivation tests pass;
- generated examples and tfplugindocs output are current and idempotent;
- no temporary `replace` directive remains; and
- formatting, lint, build, unit, race, focused acceptance, full acceptance,
  module tidy, generation, and diff checks all pass.

Until these conditions are met, the resource remains WIP and must not be used
as evidence that trusted tool modifiers are available through Terraform.
