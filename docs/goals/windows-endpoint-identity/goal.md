# Windows Endpoint Identity

## Objective

Implement the next Goal 7 compatibility-layer slice: normalize and document
Windows endpoint identity without changing Bumblebee's public record schema or
turning Windows support into a divergent scanner.

The intended outcome is a small, verified implementation that keeps
`endpoint.device_id` as the stable machine correlation key, treats
`endpoint.username` and `endpoint.uid` as scanner-process identity, and makes
Windows `uid` semantics explicit.

## Oracle

The goal is complete only when:

- The current implementation is checked against the researched identity plan.
- Windows `endpoint.uid` behavior is deliberate and does not emit a misleading
  fake Unix UID if Windows user lookup fails.
- Tests cover stable Windows endpoint identity behavior without weakening
  macOS/Linux behavior.
- `windows.md` Goal 7 is updated with the compatibility-layer decision and any
  remaining gaps.
- No JSON schema version bump or public README support expansion occurs.
- Verification passes with the local Go toolchain if `go` is still not on
  `PATH`.

## Constraints

- Preserve the existing endpoint JSON shape: `hostname`, `os`, `arch`,
  `username`, `uid`, optional `device_id`.
- Do not auto-read Windows registry, SMBIOS UUID, MachineGuid, Entra state, or
  hardware identifiers during scans.
- Do not change record identity, parser behavior, sinks, root discovery, or
  exposure matching.
- Keep the change scoped to the compatibility layer and additive docs/tests.
- Do not update README or broader user-facing support docs in this tranche.
- Prefer source-backed wording over memory for Windows identity semantics.

## Existing Plan Facts

- Go `os/user.Current()` is already the endpoint identity source.
- On Windows, Go's user lookup returns a SAM-compatible username and a user SID
  string for `User.Uid`.
- `endpoint.device_id` is already supplied only through `--device-id-env`.
- The preferred Windows `device_id` source should be an operator-provisioned
  MDM/RMM/EDR/Entra/Intune-style asset identifier, not a Bumblebee-collected
  hardware or registry fingerprint.
- The local verified Go executable is:
  `C:\Users\bbutner\AppData\Local\Programs\bumblebee-tools\go1.26.3\go\bin\go.exe`.
- Goal 7 in `windows.md` currently remains open.

## Likely Misfire

The dangerous failure mode is turning endpoint identity into a Windows-only
fingerprinting feature or schema fork. The slice should clarify and test the
existing cross-platform contract, not invent new identity fields or collect
machine IDs automatically.
