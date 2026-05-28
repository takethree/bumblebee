# T008 Live Deploy And Smoke

## Result

Done.

## Summary

Applied the pending remote D1 migration, deployed Hive, and ran a redacted live smoke against `https://hive.take3tech.dev`.

The live smoke:

- enrolled a disposable `environment=test` device through the live Hive join path;
- synced the active Hive catalog without publishing a test-only catalog;
- generated a test fixture from an active catalog entry;
- ran `bumblebee hive run` against live Hive;
- observed the resulting finding only in `environment=test`;
- verified the same test device did not appear in default production findings;
- verified default production findings were clean;
- dry-ran and confirmed the guarded device purge;
- verified the test finding was removed and the purged test device returned `404`;
- verified no test devices remained after cleanup.

## Verification

- `cd F:/bumblebee-hive; npx wrangler d1 migrations list bumblebee-hive --remote`: pass, `0010_device_environment.sql` pending before apply.
- `cd F:/bumblebee-hive; npx wrangler d1 migrations apply bumblebee-hive --remote`: pass, applied `0010_device_environment.sql`.
- `cd F:/bumblebee-hive; npx wrangler deploy`: pass, deployed Worker version `ba6db10c-d6b6-49ee-ac2d-6953260ea2fa`.
- `cd F:/bumblebee-hive; npx wrangler d1 migrations list bumblebee-hive --remote`: pass, no migrations to apply.
- Redacted live smoke: pass.
- Post-cleanup live check: pass, `test_device_total=0`, `default_production_finding_total=0`.

## Redacted Live Smoke Result

```json
{
  "ok": true,
  "hive_base_url": "https://hive.take3tech.dev",
  "device_prefix": "f7de0859",
  "environment": "test",
  "preexisting_test_devices_purged": 0,
  "test_finding_observed": true,
  "test_device_default_production_finding_total": 0,
  "default_production_finding_total": 0,
  "dry_run_response_present": true,
  "purge_ok": true,
  "post_purge_test_finding_total": 0,
  "post_purge_device_status": 404
}
```

## Notes

No secrets, raw object keys, full device IDs, local usernames, hostnames, SIDs, or full local paths are recorded here. The temporary ignored live-smoke script was removed after execution.

The overall goal is not complete yet because T009 commit/push and T999 final audit remain.
