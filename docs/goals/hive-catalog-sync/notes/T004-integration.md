# T004 Integration Receipt

## Result

Cross-repo local integration proof passed.

## Proof Path

The e2e test in `F:/bumblebee-hive/test/e2e.test.ts` now exercises the full local path:

1. Builds the current Bumblebee CLI from `F:/bumblebee`.
2. Starts a local HTTP server that forwards requests into the Hive Worker.
3. Runs `bumblebee hive join` against the local Hive `/v1/enroll` route.
4. Publishes a validated exposure catalog through `POST /v1/admin/catalog/current`.
5. Runs `bumblebee hive catalog sync` against `GET /v1/catalog/current`.
6. Runs `bumblebee hive run`, which syncs, scans a fixture project, emits one catalog finding, and uploads to Hive.
7. Runs Hive normalization on the queued batch.
8. Verifies `GET /v1/admin/findings?catalog_id=advisory-left-pad` returns the resulting critical finding.
9. Verifies the stored `scan_summary` has `status=complete`, `findings_emitted=1`, and Hive catalog metadata.

## Verification

- `BUMBLEBEE_E2E=1 BUMBLEBEE_REPO=F:/bumblebee GO_EXE=%LOCALAPPDATA%/CodexTools/go1.26.3/go/bin/go.exe npm run test:e2e` passed: 1 test file, 3 tests.
- `go test -count=1 ./...` passed in `F:/bumblebee`.
- `npm test` passed in `F:/bumblebee-hive`: 3 test files, 53 tests.
- `npm run build` passed in `F:/bumblebee-hive`.
- `git diff --check` passed in both repositories, with only CRLF working-tree warnings.

## Notes

- No production deploy, remote migration, or real credentials were used.
- No raw device IDs, secrets, local usernames, hostnames, SIDs, or profile paths were printed in the receipt.
- The full e2e path uses local fixture identifiers only.
