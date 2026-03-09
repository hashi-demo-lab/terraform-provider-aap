# 009 - Mirror Upstream Issues

**Date**: 2026-03-09
**Type**: Project Management

## Summary

Mirrored all 15 open issues from the upstream repository (`ansible/terraform-provider-aap`) to the local repository (`hashi-demo-lab/terraform-provider-aap`). Each mirrored issue includes a reference link back to the upstream original.

## Note

`hashi-demo-lab/terraform-provider-aap` is an unlinked fork of `ansible/terraform-provider-aap` — it was created as a standalone copy rather than a GitHub fork, so there is no formal fork relationship between the two repositories.

## Issue Mapping

| Local | Upstream | Title | Label |
|------:|---------:|-------|-------|
| #1 | #181 | `TestAccEDAEventStreamDataSourceRetrievesPostURL` test fails | bug |
| #2 | #173 | Feature Request: job_template and workflow_job_template resources | enhancement |
| #3 | #159 | Feature Request: Windows binaries | enhancement |
| #4 | #137 | Feature Request: Support `when = "destroy"` for aap_job | enhancement |
| #5 | #130 | Add support for limit while invoking a job in AAP | enhancement |
| #6 | #126 | aap_job resource succeeds when underlying Ansible job fails | bug |
| #7 | #125 | Add support for passing credentials on launch in aap_job | enhancement |
| #8 | #83 | Wait for completion for aap_workflow_job | enhancement |
| #9 | #79 | Feature Requests for additional resources | enhancement |
| #10 | #52 | Add a resource to sync inventory | enhancement |
| #11 | #39 | Allow token based authentication | enhancement |
| #12 | #36 | Set a Default Organization ID at the Provider Level | enhancement |
| #13 | #35 | New Resources - Any of the available in the web interface | enhancement |
| #14 | #31 | Failed to update job resource with inventory Id | bug |
| #15 | #28 | Question - Return of job status options | enhancement |

## Breakdown by Label

| Label | Count |
|-------|------:|
| enhancement | 12 |
| bug | 3 |
| **Total** | **15** |

## Convention

Each mirrored issue body begins with:

```
> Mirrored from upstream: https://github.com/ansible/terraform-provider-aap/issues/{number}
```
