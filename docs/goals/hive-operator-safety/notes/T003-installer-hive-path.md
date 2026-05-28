# T003 Installer Hive Path

## Result

Done.

## Summary

Updated the Windows installer and pilot verifier to use the Bumblebee Hive compatibility-layer path:

- installer now enrolls through `bumblebee hive join`;
- installer writes Hive `config.json` and `secrets.json`;
- installer wrapper runs `bumblebee hive run` with explicit config/cache roots;
- verifier checks the new config/secrets shape, cache root, environment, and wrapper contract;
- installer E2E now performs real enrollment, proves rerun identity reuse, publishes a catalog, and runs the generated wrapper through Hive.

## Verification

- `PowerShell parser check for scripts/install-bumblebee.ps1`: pass
- `PowerShell parser check for scripts/verify-bumblebee-pilot.ps1`: pass
- `cd F:/bumblebee-hive; npm test`: pass, 60 tests
- `cd F:/bumblebee-hive; npm run build`: pass
- `cd F:/bumblebee-hive; BUMBLEBEE_E2E=1 npm run test:e2e`: pass, 3 tests

## Notes

The first E2E attempt exposed a test harness issue: the local HTTP bridge attached an empty body to GET requests, which made catalog sync fail before the wrapper could prove `hive run`. The harness now omits bodies on empty requests, matching real HTTP behavior.

The overall goal is not complete yet. Live Hive deploy/migration and redacted live smoke verification remain queued as T008.
