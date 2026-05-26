# T001 Source Validation

## Result

The planned architecture is viable with one required Bumblebee transport change: generic environment-backed HTTP headers. Bumblebee must not gain Cloudflare-specific behavior, but Cloudflare Access service-token protection for `/v1/ingest` requires sending `CF-Access-Client-Id` and `CF-Access-Client-Secret` or an equivalent Access token header. The current HTTP sink only emits `Content-Type`, optional `Content-Encoding`, optional `User-Agent`, `Authorization`, and HMAC headers.

## Repo Evidence

- `docs/transport.md` defines the current wire contract: POST NDJSON, line-count batching, HTTPS required for non-loopback, bearer or HMAC credentials from environment variables, HMAC over exact raw POST body, gzip before HMAC, receiver verifies HMAC before decompression, and 2xx only after full durable acceptance.
- `internal/output/httpsink.go` implements bearer, HMAC, gzip, batching, HTTPS enforcement, and non-2xx failure behavior.
- `internal/output/httpsink_test.go` already pins HMAC signing, gzip body encoding, and gzip-plus-HMAC signing over compressed bytes.
- `docs/deployment-windows.md` establishes the Windows runner model: no resident service, Task Scheduler or fleet tooling, wrapper scripts, stable paths under `C:\Program Files\Bumblebee\` and `C:\ProgramData\Bumblebee\`, secrets kept out of command lines, and success defined as zero exit plus a delivered complete `scan_summary`.
- `docs/state-model.md` supports Hive's first storage scope: raw records/runs, current-state promotion only after `scan_summary.status=complete`, dedupe by endpoint/run/record, and `endpoint.device_id` supplied via `--device-id-env`.

## Official Source Findings

- Cloudflare Access supports protecting self-hosted applications, including Workers routes, and service tokens are the machine-to-machine credential shape for non-browser clients.
- Cloudflare Access service-token auth uses `CF-Access-Client-Id` and `CF-Access-Client-Secret`; Access also documents single-header service-token authentication, but that still requires a custom header shape Bumblebee cannot emit today.
- Cloudflare Workers expose request bodies through Web Fetch APIs such as `request.arrayBuffer()`, which is compatible with Bumblebee's raw-body HMAC requirement.
- Cloudflare Workers limits make bounded batch size and gzip important. Hive should reject oversized requests before doing expensive parsing and Bumblebee should keep configurable `--http-batch-size`.
- R2 is the right raw batch store for v1. Hive can write each accepted batch as an immutable object after auth verification.
- D1 is the right small relational index for devices, batches, and runs. Do not put large raw NDJSON payloads in D1.
- Queues are appropriate for post-accept normalization. They should not be required before the receiver returns 2xx unless the implementation treats queue enqueue as part of durable acceptance.
- Microsoft PowerShell supports `Get-FileHash` for checksum verification, `Get-AuthenticodeSignature` for signing validation where available, `Register-ScheduledTask` for scheduled execution, and DPAPI-backed `Export-Clixml` behavior for secure local secret storage on Windows.

Sources:

- Cloudflare Access application types: https://developers.cloudflare.com/cloudflare-one/access-controls/applications/choose-application-type/
- Cloudflare Access service tokens: https://developers.cloudflare.com/cloudflare-one/access-controls/service-credentials/service-tokens/
- Cloudflare Workers limits: https://developers.cloudflare.com/workers/platform/limits/
- Cloudflare Workers request body API: https://developers.cloudflare.com/workers/runtime-apis/request/
- Cloudflare R2 Workers API: https://developers.cloudflare.com/r2/api/workers/workers-api-reference/
- Cloudflare D1 Workers API: https://developers.cloudflare.com/d1/worker-api/
- Cloudflare Queues: https://developers.cloudflare.com/queues/
- Cloudflare Workers testing: https://developers.cloudflare.com/workers/testing/
- Microsoft Get-FileHash: https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/get-filehash
- Microsoft Get-AuthenticodeSignature: https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.security/get-authenticodesignature
- Microsoft Export-Clixml: https://learn.microsoft.com/en-us/powershell/module/microsoft.powershell.utility/export-clixml
- Microsoft Register-ScheduledTask: https://learn.microsoft.com/en-us/powershell/module/scheduledtasks/register-scheduledtask

## Recommendation

Start with the Bumblebee transport slice before creating Hive:

- Add repeatable generic `--http-header-env Header-Name=ENV_VAR`.
- Resolve header values from environment variables only.
- Reject empty values.
- Reject reserved headers managed by Bumblebee: `Content-Type`, `Content-Encoding`, `User-Agent`, `Authorization`, `X-Inventory-Signature`, and `X-Inventory-Timestamp`.
- Apply custom headers before auth, then let built-in auth set managed auth headers.
- Document the feature generically as custom receiver headers, not Cloudflare support.
- Verify with unit tests and a loopback smoke that custom headers can coexist with HMAC plus gzip.

This keeps the fork maintainable and enables Cloudflare Access without baking Cloudflare into Bumblebee.
