# T005 Hive Join Idempotency Receipt

Result: done

Implemented Bumblebee-side Hive join safety controls:

- `bumblebee hive join` now reuses a valid local Hive config by default instead of calling `/v1/enroll` again.
- Reused joins preserve the existing device ID, HMAC key, environment, scan profile, and roots unless compatible updates are explicitly provided.
- First join or `--new-device` performs enrollment.
- Enrollment accepts `--environment production|test`; an existing device cannot be silently moved to another environment without `--new-device`.
- Reused joins no longer require the enrollment token, which supports idempotent installer reruns.
- Hive enrollment requests now send the environment in the JSON body and store the returned/default environment in local config.
- Cross-repo E2E now proves `--environment test`, idempotent join reuse, explicit `--new-device`, and production-default finding isolation.
- README and transport docs describe Hive-managed transport, idempotent join, and environment behavior.

Verification:

- `cd F:/bumblebee; $env:GOEXPERIMENT=''; & $env:LOCALAPPDATA\CodexTools\go1.26.3\go\bin\go.exe test -count=1 ./...` passed.
- `cd F:/bumblebee-hive; $env:BUMBLEBEE_E2E='1'; $env:BUMBLEBEE_REPO='F:\bumblebee'; $env:GO_EXE="$env:LOCALAPPDATA\CodexTools\go1.26.3\go\bin\go.exe"; npm run test:e2e` passed.
- `cd F:/bumblebee-hive; npm test` passed.
- `cd F:/bumblebee-hive; npm run build` passed.

Residual notes:

- Windows installer/verification scripts still need to call the new Hive compatibility-layer path; that remains T003.
- Hive purge is not implemented yet; that remains T006.
