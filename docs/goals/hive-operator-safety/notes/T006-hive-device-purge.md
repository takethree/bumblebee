# T006 Hive Device Purge Receipt

Result: done

Implemented guarded Hive device purge:

- Added `POST /v1/admin/devices/<device-id>/purge?dry_run=true|false`.
- Dry-run returns aggregate candidate counts only.
- Confirmed purge requires a non-empty reason and matching `confirm_device_id`.
- Production devices must be disabled before destructive purge; test devices may be purged directly.
- Purge deletes matching raw R2 objects first, then D1 metadata for the device, lifecycle events, runs, batches, normalization jobs, inventory records/current rows, exposure findings, and the device row.
- If any raw object delete fails, Hive returns `ok:false` and leaves D1 metadata in place for retry.
- Added `scripts/invoke-device-purge.ps1` with dry-run default and explicit `-ConfirmPurge`.
- README and developer rollout runbook document the operator workflow and safety guards.
- Tests prove dry-run counts, confirmation mismatch, confirmed purge, production active-device rejection, no raw key exposure, and D1/R2 cleanup.

Verification:

- `cd F:/bumblebee-hive; npm test` passed: 3 files, 60 tests.
- `cd F:/bumblebee-hive; npm run build` passed.
- PowerShell parser check for `scripts/invoke-device-purge.ps1` passed.

Residual notes:

- Installer/verification script update remains T003.
- Full local cross-repo verification remains T007.
- Live deploy/migrate/smoke remains T008.
