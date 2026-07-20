# AGENTS.md

Guidance for AI coding agents working in the `oci-delta` repository in the `containers` organization.

## Project Overview

`oci-delta` is a Go project and a CLI toolset for taking the diff between two OCI Images and producing a
delta that can be used to deploy updates to bootc systems with the older of the two OCI Images.

Key dependencies include `containers/image`, `containers/storage`, sigstore, and `containers/tar-diff`.
Other dependencies can be found within [go.mod](go.mod).

Please reference [README.md](README.md) if more details or context is needed.

## Agent Workflow Guidelines

**For small and/or simple changes:** Propose a short plan and ask a human to review it before you
start implementing.

**Never assume or guess** intent or meaning. If instructions, code, or documentation are ambiguous,
ask for clarification.

**If you find a potential bug** in the code, inform the developer.

**If you find a potential security issue** in the code, inform the developer and reference the
guidelines in [SECURITY.md](SECURITY.md).

**Follow development conventions**: All code style, commit formatting, and testing requirements
are in [CONTRIBUTING.md](CONTRIBUTING.md).

## AI Permissions

**Always allowed**: Read any file, run formatting/linting commands (`make fmt`, `make lint`), search
GitHub issues and PRs, analyze logs and test output.

**Safe to change**: Log messages, comments, unexported or local variable names, code formatting.

**Ask first**: Anything that affects core algorithms, delta file format, or cross-platform compatibility.
This includes the following list:

- Documentation files (`README.md`, `CONTRIBUTING.md`, `AGENTS.md`).
- Test assertions.
- Linting config (`.golangci.yml`).
- Refactoring.
- Core delta create/apply/import logic (`pkg/oci-delta/`).
- Delta file format specification (in `README.md`).
- Build configuration (`Makefile`, `.packit.yaml`).
- CI workflows (`.github/workflows/`).
- Dependency changes (`go.mod`).
- Changes to exported APIs.

**Never**: Push directly to `main`, commit compiled binaries or files produced during testing, break
cross-platform compatibility, log or hardcode sensitive information.

## Platform Requirements

The target systems are Linux (x86_64 and ARM) and has to run bootc version 1.15.0 or later.
Reference [README.md](README.md) if more details are needed.
