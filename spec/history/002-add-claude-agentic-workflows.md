# 002 - Add Claude Agentic Workflows

**Date**: 2026-03-09

## Summary

Copied the `.claude` directory from `hashi-demo-lab/terraform-agentic-workflows` into this repository to enable Claude Code agentic workflows for Terraform provider development.

## Why

The `terraform-agentic-workflows` repo contains a curated set of Claude Code agents and skills tailored for Terraform provider, module, and consumer development. Adding these to this provider repo gives contributors access to specialized agents for design, development, research, testing, and validation workflows.

## What Changed

### Added `.claude/` directory (103 files)

- **`CLAUDE.md`** — Project-level instructions for Claude Code
- **`settings.local.json`** — Local Claude Code settings
- **`agents/`** — 14 agent definitions covering provider, module, and consumer workflows (design, development, research, testing, validation)
- **`skills/`** — 25 skill directories covering implementation patterns, test patterns, security baselines, style guides, and more

## Source

```
https://github.com/hashi-demo-lab/terraform-agentic-workflows/tree/main/.claude
```
