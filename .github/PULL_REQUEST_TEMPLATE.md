<!--
  Thanks for contributing to terraform-provider-anypoint.
  Please fill in the sections below.
-->

## Summary

<!-- 1–3 bullets. What does this PR do, at a glance? -->

-
-

## Related

- Closes: #<issue>
- Generator PR (if a spec change was required): <link>
- `anypoint-client-go/<service>` release tag: `<service>/vX.Y.Z`

## Type of change

- [ ] New resource
- [ ] New data source
- [ ] Bug fix in existing resource / data source
- [ ] Documentation only
- [ ] Dependency / client module bump
- [ ] Refactor / internal change

## Checklist

### Code

- [ ] Resource / data source files follow `AGENTS.md` skeleton (CRUD signatures, `get<Name>AuthCtx`, structured `diag.Diagnostic`).
- [ ] `RESOURCES_MAP` and/or `DATASOURCES_MAP` updated in `anypoint/provider_resources.go` / `anypoint/provider_datasources.go`.
- [ ] Immutable fields marked `ForceNew: true`.
- [ ] Composite IDs use `COMPOSITE_ID_SEPARATOR` via `ComposeResourceId` / `DecomposeResourceId`.
- [ ] `Importer` block present if resource is importable.
- [ ] `gofmt -w anypoint/` and `goimports -w anypoint/` clean.
- [ ] `make build` passes.

### Spec / client dependency

- [ ] If this PR required an OAS spec change, the matching `anypoint-automation-client-generator` PR is merged to `master` AND the affected client module has been released with a version tag.
- [ ] `go.mod` has **no** `replace` directive pointing at a local path for `anypoint-client-go/*`. Verified with `grep -n "replace " go.mod`.
- [ ] `go.sum` updated via `go mod tidy`.

### Docs & examples

- [ ] `examples/resources/anypoint_<name>/{resource.tf, import.sh}` exists.
- [ ] `examples/data-sources/anypoint_<name>/data-source.tf` exists.
- [ ] `tfplugindocs generate` ran; generated `docs/` committed.

### Verification

- [ ] `make install` against local provider.
- [ ] Playground `case_<N>.tf` apply → state matches API.
- [ ] Update path tested (each updatable field exercised independently).
- [ ] Drift check: `terraform plan` after apply shows no diff.
- [ ] `terraform destroy` clean; remote API confirms removal.

## Notes for reviewer

<!-- Anything reviewers should know: regions exercised, HAR evidence, manual API quirks, etc. -->
