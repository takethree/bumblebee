# Transport

`bumblebee` writes NDJSON to one of three sinks:

1. `stdout`
2. local file (`--output file --output-file PATH`)
3. HTTP(S) POST (`--output http --http-url URL`)

The collector does not spool, retry in the background, or keep an on-disk
queue. One invocation runs, writes, emits diagnostics and an optional
`scan_summary`, then exits.

## Choosing a sink

### stdout

Best for local testing and one-off remote execution:

```sh
bumblebee scan --profile deep --root "$HOME" > inventory.ndjson
```

Diagnostics are written to stderr as NDJSON.

### file

Use this when the endpoint already has a log shipper:

```sh
bumblebee scan \
  --profile baseline \
  --output file \
  --output-file /var/log/bumblebee/inventory.ndjson \
  --append
```

The file sink is plain NDJSON. Compress or ship downstream if needed.

### http

Use this when there is no existing endpoint log pipeline:

```sh
bumblebee scan \
  --profile deep \
  --root "$HOME" \
  --exposure-catalog ./catalog.json \
  --output http \
  --http-url https://inventory.example.com/v1/ingest \
  --http-auth bearer \
  --http-token-env BUMBLEBEE_TOKEN \
  --http-gzip \
  --device-id-env BUMBLEBEE_DEVICE_ID
```

HTTP behavior:

- Each POST body is NDJSON with `Content-Type: application/x-ndjson`.
- Batching is line-count based. `--http-batch-size` controls records per POST.
- HTTPS is required for non-loopback hosts unless `--http-allow-insecure` is set.
- Bearer tokens and HMAC keys are read from environment variables, not CLI literals.
- Additional static headers can be supplied with repeatable
  `--http-header-env Header-Name=ENV_VAR`; values are read from environment
  variables so receiver credentials do not appear as command-line literals.
- Non-2xx responses fail the run immediately.
- `scan_summary.http_batches_attempted`, `http_batches_succeeded`,
  `http_batches_failed`, and `http_last_status` report delivery results seen
  by the sink.

### Hive-managed transport

`bumblebee hive run` is a compatibility layer over the normal HTTP sink. It
loads Hive config written by `bumblebee hive join`, syncs the current Hive
catalog, then invokes the same scan transport with HMAC auth, Cloudflare Access
service-token headers, gzip, and the saved Hive device ID.

`bumblebee hive join` reuses a valid local config by default so rerunning an
installer does not create duplicate Hive devices. It enrolls only when no
config exists or when `--new-device` is passed. New enrollments can be marked
with `--environment production|test`; Hive stores that label and defaults
operator views to production data.

## Exact HTTP wire contract

Every request uses:

- Method: `POST`
- Header: `Content-Type: application/x-ndjson`
- Header: `User-Agent: bumblebee/<scanner_version>` when available
- Body: one JSON object per line, no wrapper object, no JSON array

Optional auth headers:

- Bearer mode: `Authorization: Bearer <token>`
- HMAC mode: `X-Inventory-Signature: sha256=<hex>`
- HMAC mode with timestamp: `X-Inventory-Timestamp: <unix-seconds>`

Optional custom headers:

- `--http-header-env Header-Name=ENV_VAR` adds one static request header
  whose value is read from `ENV_VAR`.
- Custom headers cannot override headers managed by Bumblebee or the HTTP
  client, including `Content-Type`, `Content-Encoding`, `Content-Length`,
  `Host`, `User-Agent`, `Authorization`, `X-Inventory-Signature`, and
  `X-Inventory-Timestamp`.
- Use custom headers for receiver-side routing or gateway credentials that are
  independent of Bumblebee's bearer/HMAC auth.

HMAC signing rules:

- Without `X-Inventory-Timestamp`, the signature input is the raw POST body.
- With `X-Inventory-Timestamp`, the signature input is
  `<timestamp>.<raw-post-body>`.
- The signature is hex-encoded SHA-256 HMAC and is sent as
  `sha256=<hex>`.

Gzip rules:

- When `--http-gzip` is set, the request also carries
  `Content-Encoding: gzip`.
- Compression happens before HMAC signing.
- Receivers that verify HMAC must verify the signature against the exact raw
  received bytes, then decompress if `Content-Encoding: gzip` is present.

## Receiver behavior

A receiver should:

1. Accept `POST` requests whose `Content-Type` is `application/x-ndjson`.
2. Read the raw request body exactly as received.
3. If HMAC is enabled, verify the signature before decompression.
4. If `Content-Encoding: gzip` is present, decompress only after signature
   verification.
5. Treat the payload as newline-delimited JSON objects.
6. Return `2xx` only after the full batch is durably accepted.
7. Return non-2xx for any batch that was not durably accepted in full.

Partial acceptance is not part of the protocol. One request is one batch:
either the entire batch is accepted and acknowledged with `2xx`, or the
sender treats it as failed.

## Record stream semantics

The stream can contain these `record_type` values:

- `package`
- `finding`
- `scan_summary`
- `diagnostic` on stderr only, not in the records sink

`record_id` is stable for the record's canonical identity within a run and is
intended for receiver-side dedupe. `run_id` stays separate and identifies the
invocation.

`--findings-only` suppresses only `record_type=package`. Findings,
`scan_summary`, and diagnostics still flow.

## `scan_summary` completion semantics

Receivers should promote a run to current state only after a matching
`scan_summary` with `status=complete`.

Interpretation of summary status:

- `complete`: the run completed and the sender observed no terminal scan or
  sink errors.
- `partial`: the run emitted some records but also hit a terminal scan error
  or sink delivery failure.
- `error`: the run failed before producing usable package state.

Important sink nuance:

- Failed HTTP POST batches must not be treated as a successful complete run.
- If the HTTP sink reports failed batches, the run is not a trustworthy
  complete snapshot even if package parsing otherwise succeeded.
- `http_last_status=0` means the last batch failed before an HTTP response was
  received.

For `--findings-only` runs, `scan_summary.package_records_emitted` can be `0`
while `package_records_suppressed` is positive. That still reflects a valid
run shape; it simply means package records were intentionally withheld from the
records sink.

Hive-managed runs may include optional `scan_summary.catalog` metadata. This
metadata records the Hive catalog release id, bundle hash, source label,
publish/sync times, and entry count used for that run. Normal `bumblebee scan`
does not fetch Hive data or add catalog metadata unless a Hive command invokes
the scan path with a synced catalog.

## Why no direct object-storage sink

Direct S3/GCS/Azure Blob upload from endpoints is intentionally out of scope
for v0.1. It would push cloud credentials or presigned URL management onto
every endpoint and create one object per endpoint run.

If object storage is the final destination, send to a small internal HTTP
relay and let the relay batch and write centrally.
