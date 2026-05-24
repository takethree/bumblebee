# Windows HTTP Sink Smoke

## Objective

Implement and verify the Windows HTTP sink smoke slice from `windows.md` Goal 12.

The intended outcome is a deterministic local Windows smoke check proving that
`bumblebee.exe` can deliver NDJSON package records and a complete
`scan_summary` through `--output http` to a loopback receiver, without using an
external ingest service or writing secrets into repo artifacts.

## Oracle

The goal is complete only when:

- `scripts/windows-smoke.ps1` runs on Windows and includes an HTTP sink smoke
  against a local loopback receiver.
- The redacted smoke summary proves the HTTP receiver got package records and a
  `scan_summary` with `status=complete`.
- Bearer auth is supplied through an environment variable, verified by the
  receiver, and not written to the redacted summary.
- `windows.md` Goal 12 is updated with a receipt and the HTTP sink checkbox is
  checked.
- Verification passes:
  - `go test ./cmd/bumblebee ./internal/...`
  - `powershell -ExecutionPolicy Bypass -File scripts\windows-smoke.ps1`
  - `git diff --check -- scripts/windows-smoke.ps1 windows.md`

## Constraints

- Keep this as a Windows compatibility-layer validation step.
- Do not add or change product code unless the smoke exposes a real CLI/sink bug.
- Use a local loopback receiver; do not depend on an external HTTP service.
- Keep raw NDJSON and received request bodies under the temp evidence directory.
- Do not store bearer tokens, HMAC keys, or raw inventory data in repo files.
- Do not expand scope into HMAC, gzip, deployment docs, or native Windows
  ecosystem support in this tranche.

## Existing Plan Facts

- Extend `scripts/windows-smoke.ps1` with a loopback HTTP receiver using
  PowerShell/.NET facilities rather than a third-party dependency.
- Generate a small fixture project under the temp evidence directory with a
  `package-lock.json`.
- Run `bumblebee.exe scan --profile project --root <fixture> --output http`
  against the local receiver with `--http-auth bearer`,
  `--http-token-env BUMBLEBEE_SMOKE_HTTP_TOKEN`, and `--http-batch-size 1`.
- Fail the smoke if no HTTP request is received, bearer auth is wrong, no package
  record arrives, no `scan_summary` arrives, or the summary is not complete.
- Add a redacted `http_sink` section to `smoke-summary.redacted.json`.
- Update `windows.md` with the passing receipt.

## Likely Misfire

The main failure mode is treating this as a generic HTTP sink refactor or
deployment task. The goal is narrower: prove the existing HTTP sink path works
from the Windows binary in the smoke workflow.
