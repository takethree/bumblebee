# T004 Hive Environment Isolation Receipt

Result: done

Implemented durable Hive device environment isolation for operator safety:

- Added `devices.environment` migration with `production` default and `production|test` check.
- `POST /v1/enroll` now accepts optional `environment` and defaults omitted values to `production`.
- Admin metadata routes default to `environment=production` and accept explicit `environment=production|test|all` where list/overview/health/attention/package visibility is environment-scoped.
- Admin UI has a recoverable environment selector; non-production selections are encoded in the URL and sent to admin metadata requests.
- Device list/detail responses expose only the environment label needed for operator context.
- README documents the production-default behavior and environment query parameter.
- Tests cover enrollment defaults/rejection, default production isolation, explicit test/all views, URL state, and UI request propagation.

Verification:

- `cd F:/bumblebee-hive; npm test` passed: 3 files, 58 tests.
- `cd F:/bumblebee-hive; npm run build` passed: `tsc --noEmit`.

Residual notes:

- Bumblebee-side `hive join` does not send `environment` yet; that remains T005.
- Installer/verification scripts still need to use the joined Hive compatibility-layer path; that remains T003.
- Purge is not implemented yet; that remains T006.
