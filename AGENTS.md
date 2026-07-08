# AGENTS.md

Agent rules for `terraform-provider-anypoint`. Read by Claude Code (via `CLAUDE.md` import), Roo Code, Cursor, and any tool that supports the `AGENTS.md` convention.

## Project context

- Terraform provider for MuleSoft Anypoint Platform.
- Language: Go 1.20+. Framework: `terraform-plugin-sdk/v2`.
- Client libs: `github.com/mulesoft-anypoint/anypoint-client-go/*` — one Go module per Anypoint API, generated from OAS3 specs maintained in the sibling repo `anypoint-automation-client-generator`.
- Provider source address used in user Terraform configs: `anypoint.mulesoft.com/automation/anypoint`.

## Code layout

Single Go package `anypoint/` (no subpackages):

```
anypoint/
  provider.go              # Provider() schema, auth, providerConfigure
  provider_clients.go      # ProviderConfOutput: typed API clients per submodule
  provider_resources.go    # RESOURCES_MAP   (register new resources here)
  provider_datasources.go  # DATASOURCES_MAP (register new data sources here)
  resource_*.go            # CRUD implementations
  data_source_*.go         # Read implementations (often plural variant *_xxxs.go)
  util.go                  # type helpers, COMPOSITE_ID_SEPARATOR
main.go                    # plugin.Serve entry point
```

`meta` passed to every CRUD function is a `ProviderConfOutput` value (defined in `provider_clients.go`). Pull the typed client off it (`pco.<service>client`). Do not construct API clients inside resource files.

## Resource / data source skeleton

```go
func resource<Name>() *schema.Resource {
    return &schema.Resource{
        CreateContext: resource<Name>Create,
        ReadContext:   resource<Name>Read,
        UpdateContext: resource<Name>Update,
        DeleteContext: resource<Name>Delete,
        Description:   "Creates and manages a `<name>` component.",
        Schema: map[string]*schema.Schema{
            "last_updated": { Type: schema.TypeString, Optional: true, Computed: true,
                Description: "The last time this resource has been updated locally." },
            "id": { Type: schema.TypeString, Computed: true,
                Description: "The unique id of this resource." },
            // ...
        },
    }
}
```

CRUD signatures:

```go
func resource<Name>Create(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics
func resource<Name>Read  (ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics
func resource<Name>Update(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics
func resource<Name>Delete(ctx context.Context, d *schema.ResourceData, m any) diag.Diagnostics
```

Data source skeleton mirrors the resource one but only `ReadContext` is set, and `id` is `Required: true` for singular lookup or `Computed: true` for a listing variant.

## Provider configuration access

`meta` arrives as `ProviderConfOutput` by value:

```go
pco := m.(ProviderConfOutput)
```

The typed clients are on `pco` as unexported fields named `<service>client` (e.g. `pco.vpcclient`, `pco.dlbclient`, `pco.apimclient`). Field names follow the `provider_clients.go` `ProviderConfOutput` struct definition.

## Auth context

Each resource file defines its own `get<Name>AuthCtx` helper that wraps `ctx` with the access token and server index from `pco`, using the **target service's** own `ContextAccessToken` / `ContextServerIndex` constants:

```go
func getVPCAuthCtx(ctx context.Context, pco *ProviderConfOutput) context.Context {
    tmp := context.WithValue(ctx, vpc.ContextAccessToken, pco.access_token)
    return context.WithValue(tmp, vpc.ContextServerIndex, pco.server_index)
}
```

Call pattern inside CRUD:

```go
authctx := getVPCAuthCtx(ctx, &pco)
res, httpr, err := pco.vpcclient.DefaultApi.OrganizationsOrgIdVpcsPost(authctx, orgid).VpcCore(*body).Execute()
```

`pco.access_token` and `pco.server_index` are unexported — only auth-ctx helpers in `anypoint/` may read them.

## Error handling

The codebase uses **structured `diag.Diagnostic`** throughout — not `diag.FromErr`. Standard pattern:

```go
res, httpr, err := pco.vpcclient.DefaultApi.Op(authctx, orgid, id).Execute()
if err != nil {
    var details string
    if httpr != nil && httpr.StatusCode >= 400 {
        defer httpr.Body.Close()
        b, _ := io.ReadAll(httpr.Body)
        details = string(b)
    } else {
        details = err.Error()
    }
    diags := append(diags, diag.Diagnostic{
        Severity: diag.Error,
        Summary:  "Unable to <verb> <resource> " + id,
        Detail:   details,
    })
    return diags
}
defer httpr.Body.Close()
```

Summary convention: `"Unable to <verb> <resource> <id-or-name>"` — capital `U`, present tense verb (create, read, update, delete, patch, list, move). Detail field holds the raw API body or `err.Error()`.

## Schema field conventions

| Attribute | Use when | Example |
|---|---|---|
| `Required: true` | User must provide | `org_id`, `name` |
| `Optional: true` | User may provide | `description`, `tags` |
| `Computed: true` | API sets it | `id`, `created_at`, `status` |
| `ForceNew: true` | Change implies replacement | `region`, `cidr_block`, `org_id` |
| `Sensitive: true` | Mask in output | `client_secret`, `password` |

Typically `Required` + `ForceNew`: `org_id`, `env_id`, `vpc_id`, `region`, `cidr_block`.
Typically `Computed`: `id`, `created_at`, `updated_at`, `owner_id`, `status`.

## IDs

Simple ID from the platform:

```go
d.SetId(resource.GetId())
```

Composite ID — use `COMPOSITE_ID_SEPARATOR = "/"` from `util.go`:

```go
d.SetId(fmt.Sprintf("%s/%s", orgId, resource.GetId()))
```

## Avoid

- `panic()` in provider code
- Ignored errors (`_ = something()`)
- Global mutable state
- Reflection (use `util.go` helpers — `toInt32`, `ListInterface2ListStrings`, `ConvPrimtiveInterface2String`)
- Constructing API clients inside resource files — they belong on `ProviderConfOutput`
- `diag.FromErr` — codebase convention is structured `diag.Diagnostic`

---

# Anypoint domain conventions

## Resource name prefixes

| Prefix | Anypoint concept | Example |
|---|---|---|
| `ame` | Anypoint MQ Exchange | `resource_ame.go` |
| `amq` | Anypoint MQ Queue | `resource_amq.go` |
| `apim` | API Manager | `resource_apim_mule4.go` |
| `bg` | Business Group | `resource_bg.go` |
| `dlb` | Dedicated Load Balancer | `resource_dlb.go` |
| `env` | Environment | `resource_env.go` |
| `idp` | Identity Provider | `resource_idp_oidc.go` |
| `rolegroup` | Role Group | `resource_rolegroup.go` |
| `secretgroup` | Secret Group | `resource_secretgroup.go` |
| `vpc` | Virtual Private Cloud | `resource_vpc.go` |
| `vpn` | VPN | `resource_vpn.go` |

Variants use `resource_<service>_<variant>.go` (e.g. `resource_apim_policy_jwt_validation.go`).

## Authentication modes

Priority order in `providerConfigure`:

1. `access_token` (env: `ANYPOINT_ACCESS_TOKEN`)
2. `username` + `password` (env: `ANYPOINT_USERNAME` / `ANYPOINT_PASSWORD`) — **deprecated**
3. `client_id` + `client_secret` (env: `ANYPOINT_CLIENT_ID` / `ANYPOINT_CLIENT_SECRET`) — preferred connected-app flow

## Control planes

| `cplane` value | Server index | Region |
|---|---|---|
| `us` (default) | 0 | US |
| `eu` | 1 | EU |
| `gov` | 2 | GOV |

`cplane` env var: `ANYPOINT_CPLANE`. Converted by `cplane2serverindex` and propagated via per-service `ContextServerIndex`.

## Common schema fragments

### `org_id`

```go
"org_id": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "The organization id where the resource is defined.",
}
```

### `env_id`

```go
"env_id": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "The environment id where the resource is defined.",
}
```

### `region`

```go
"region": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "The CloudHub region where this resource will exist.",
}
```

### `vpc_id`

```go
"vpc_id": {
    Type:        schema.TypeString,
    Required:    true,
    ForceNew:    true,
    Description: "The VPC id where the resource is deployed.",
}
```

### `cidr_block`

```go
"cidr_block": {
    Type:             schema.TypeString,
    Required:         true,
    ForceNew:         true,
    Description:      "The IP address range (largest /16, smallest /24).",
    ValidateDiagFunc: validation.ToDiagFunc(validation.IsCIDR),
}
```

### `cplane` validation

```go
ValidateFunc: func(val any, key string) (warns []string, errs []error) {
    v := val.(string)
    if v != "us" && v != "eu" && v != "gov" {
        errs = append(errs, fmt.Errorf("%q must be 'us', 'eu', or 'gov', got: %s", key, v))
    }
    return
},
```

## Client library aliases

Conventional short aliases:

```go
import (
    vpc      "github.com/mulesoft-anypoint/anypoint-client-go/vpc"
    dlb      "github.com/mulesoft-anypoint/anypoint-client-go/dlb"
    team     "github.com/mulesoft-anypoint/anypoint-client-go/team"
    user     "github.com/mulesoft-anypoint/anypoint-client-go/user"
    env      "github.com/mulesoft-anypoint/anypoint-client-go/env"
    org      "github.com/mulesoft-anypoint/anypoint-client-go/org"
    auth     "github.com/mulesoft-anypoint/anypoint-client-go/authorization"
)
```

One alias per service. Match the submodule directory name.

---

# Workflows

## Adding a new resource

1. Confirm the client library exists. Look for `github.com/mulesoft-anypoint/anypoint-client-go/<service>` in `go.mod`. If missing, contribute an OAS3 spec to `anypoint-automation-client-generator` first, regenerate, then bump the version here.
2. Register a typed client in `anypoint/provider_clients.go`: add the module import, add a `<service>client *<service>.APIClient` field on `ProviderConfOutput`, instantiate it in `newProviderConfOutput`.
3. Create `anypoint/resource_<name>.go` with `resource<Name>{Create,Read,Update,Delete}` and a schema. Mark immutable fields `ForceNew: true`. Define `get<Name>AuthCtx` in the same file.
4. Register in `RESOURCES_MAP` (`anypoint/provider_resources.go`).
5. Create the data source `anypoint/data_source_<name>.go` if applicable. Register in `DATASOURCES_MAP` (`anypoint/provider_datasources.go`).
6. Document. Resource/datasource `Description` fields drive `tfplugindocs`; regenerate after schema changes.
7. Add a usage snippet to `examples/`.

## Adding a new data source

Same as above, minus CRUD — only `dataSource<Name>Read`. Register in `DATASOURCES_MAP`.

## Build, install, test

```bash
make build       # go build -o terraform-provider-anypoint
make install     # build + copy to ~/.terraform.d/plugins/anypoint.mulesoft.com/automation/anypoint/${VERSION}/${OS_ARCH}
make test        # go test, parallel 4, 30s timeout
make testacc     # TF_ACC=1, 120m, hits real APIs — requires ANYPOINT_CLIENT_ID / ANYPOINT_CLIENT_SECRET / ANYPOINT_ORG_ID env vars
make release     # cross-compile all platforms into ./bin/
```

Single test: `go test ./anypoint -run TestName -v`.

Edit `VERSION` and `OS_ARCH` in the `Makefile` to match your platform before `make install`. Defaults: `1.8.5-SNAPSHOT`, `darwin_arm64`.

`GOPRIVATE` must include `github.com/mulesoft-anypoint`:

```bash
go env -w GOPRIVATE=github.com/mulesoft-anypoint
```

## Debug with Delve

Terminal 1:

```bash
make debug
```

Output ends with `TF_REATTACH_PROVIDERS=...`. Copy.

Terminal 2:

```bash
TF_REATTACH_PROVIDERS='<paste>' terraform apply -var-file="params.tfvars.json"
```

User Terraform config must reference source `anypoint.mulesoft.com/automation/anypoint` for reattach to bind.

## Sample apply

`make install` first, then:

```bash
cd examples
# fill main.tf or create params.tfvars.json with client_id, client_secret, org_id
terraform init && terraform apply -var-file="params.tfvars.json"
```

## Playground

Scratch Terraform repo for end-to-end testing the locally-built provider against real Anypoint. Path is machine-specific; see `docs/MULTIREPO.md` (gitignored) for the maintainer's setup.

Cycle: `make install` here, then in the playground `terraform init -upgrade && terraform apply -var-file=params.tfvars.json`. When bumping `VERSION` in `Makefile`, bump `version =` in the playground's `main.tf` to match — otherwise Terraform won't pick up the new local plugin.

## Local client-library development

When a feature or fix requires changes to `anypoint-client-go/<service>`:

1. Edit the OAS3 spec in `anypoint-automation-client-generator/spec/<service>/`.
2. Regenerate in the generator repo:
   ```bash
   cd ../anypoint-automation-client-generator && make generate
   ```
3. Point this provider at the locally-generated module:
   ```bash
   go mod edit -replace github.com/mulesoft-anypoint/anypoint-client-go/<service>=../anypoint-automation-client-generator/dest/<service>
   go mod tidy   # picks up any new transitive deps (e.g. gopkg.in/validator.v2)
   ```
4. Build / test in the provider (`make build`, `make testacc`, `make debug`). Iterate spec → regenerate → re-test — no further `go mod` changes needed.
5. Commit + merge the generator repo. Pipeline publishes generated modules to `github.com/mulesoft-anypoint/anypoint-client-go`.
6. Manually release the affected module on `anypoint-client-go` (tagged version).
7. Switch this provider back to the released version:
   ```bash
   go mod edit -dropreplace github.com/mulesoft-anypoint/anypoint-client-go/<service>
   go get github.com/mulesoft-anypoint/anypoint-client-go/<service>@<new-version>
   go mod tidy
   ```
8. Commit `go.mod` + `go.sum`.

**Never commit a `replace` directive pointing at a local path.** Before any commit touching `go.mod`:

```bash
grep -n "replace " go.mod   # expect no anypoint-client-go entries
```

## Regenerate docs

```bash
tfplugindocs generate
```

Generated `docs/` and `templates/` are committed.

## Format

```bash
gofmt -w anypoint/
goimports -w anypoint/
```

## Clean local plugin

```bash
rm -rf ~/.terraform.d/plugins/anypoint.mulesoft.com/
```

## Release

1. Bump `VERSION` in `Makefile`.
2. `make release` — builds all platforms into `./bin/`.
3. Follow the Terraform registry publishing guide.
4. GitHub release with binaries.

## Logging

Use stdlib `log` with bracketed level prefix; the Terraform Plugin SDK routes it:

```go
log.Println("[DEBUG] something happened")
```

---

# Go style

## Principles

- Follow `gofmt` / `goimports`. Explicit error handling. No `panic()` in provider code.
- Prefer composition over inheritance. No global mutable state.

## Naming

- `resource<Name>Create / Read / Update / Delete`
- `dataSource<Name>Read`
- `get<Name>AuthCtx` (auth-context helper, per resource)
- `newProviderConfOutput` (constructor-style)
- `cplane2serverindex` (converter-style)
- Short names in narrow scope: `ctx`, `d`, `m`, `pco`
- Package-level constants in `SCREAMING_SNAKE`: `RESOURCES_MAP`, `DATASOURCES_MAP`, `COMPOSITE_ID_SEPARATOR`
- Acronyms in all caps in identifiers: `VPC`, `DLB`, `API`, `HTTP`

## Imports

Grouped: stdlib, third-party, project-internal. Example:

```go
import (
    "context"
    "fmt"
    "io"
    "sort"
    "time"

    "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
    "github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

    vpc "github.com/mulesoft-anypoint/anypoint-client-go/vpc"
)
```

One resource (or closely-related variants) per file. Every file declares `package anypoint`.

## Context

Every API call takes a derived `authctx`:

```go
authctx := get<Name>AuthCtx(ctx, &pco)
```

Never pass bare `ctx` to a client method — region routing requires `ContextServerIndex`.

## Type conversions

Safe assertions from Terraform state:

```go
orgId     := d.Get("org_id").(string)
port      := d.Get("port").(int)
isDefault := d.Get("is_default").(bool)
```

Use `util.go` helpers for list/primitive coercion (`ListInterface2ListStrings`, `ConvPrimtiveInterface2String`, `toInt32`).

## State writes

```go
d.SetId(resource.GetId())
d.Set("name", resource.GetName())
d.Set("last_updated", time.Now().Format(time.RFC850))

if err := d.Set("field", value); err != nil {
    diags = append(diags, diag.Diagnostic{
        Severity: diag.Error,
        Summary:  "Unable to set field",
        Detail:   err.Error(),
    })
    return diags
}
```

## Comments

Function comments for exported funcs. Inline comments only where the code is non-obvious — do not narrate.
