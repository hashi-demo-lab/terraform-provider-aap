# 001 - Break Upstream Fork Dependency

This is for isolated testing only, we will work with a fork once process refined...
For now this is being used for LAB testing and process validation

**Date**: 2026-03-09

## Summary

Detached this repository from the upstream fork `github.com/ansible/terraform-provider-aap` so it stands alone as an independent project under `github.com/hashi-demo-lab/terraform-provider-aap`.

## Why

This provider is maintained independently and should not carry a dependency on the upstream Ansible namespace. Breaking the fork dependency ensures:

- Go module path matches the actual repository location
- Terraform registry address reflects the correct provider namespace
- Import paths are self-consistent and do not reference the upstream org
- Git remote points to the correct origin

## What Changed

### Go module path (`go.mod`)
```
- module github.com/ansible/terraform-provider-aap
+ module github.com/hashi-demo-lab/terraform-provider-aap
```

### Go import paths (22 `.go` files)
All internal imports updated from `github.com/ansible/terraform-provider-aap/...` to `github.com/hashi-demo-lab/terraform-provider-aap/...`.

Files affected:
- `main.go`
- `internal/provider/*.go` (18 files)
- `internal/provider/customtypes/*_test.go` (2 files)

### Terraform registry address (`main.go`)
```
- Address: "registry.terraform.io/ansible/aap"
+ Address: "registry.terraform.io/hashi-demo-lab/aap"
```

### Git remote
```
- origin https://github.com/ansible/terraform-provider-aap.git
+ origin https://github.com/hashi-demo-lab/terraform-provider-aap.git
```

## Note

To fully detach the fork on GitHub (remove the "forked from" badge), an admin of `hashi-demo-lab/terraform-provider-aap` must contact GitHub Support or use repository settings to detach the fork.
