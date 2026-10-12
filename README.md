# Agent OS Indexer

Independent process that schedules knowledge indexing. Knowledge is the sole
database schema and transaction owner; this process consumes its fixed
`indexing` Go module package. It imports no Knowledge HTTP, upload, retrieval,
admin routes or migration implementation, and contains no business SQL.

## Deploy separately

Go 1.27.2. Build with `GOWORK=off GOPROXY=off go build -mod=vendor ./cmd/indexer`,
`./scripts/Build.ps1`, or the repository Dockerfile. No adjacent checkout is
required. Configure the template in `deploy/service.example.json`; real secrets
belong in a protected configuration file. Port defaults to 8097; empty Registry
and Runtime URLs retain local ports 8096 and 8094 after config loading. Container
deployments must supply explicit service URLs. Runtime receives instance
heartbeats; authenticated readiness checks only this consumer's required tables.

Provision this process with its distinct application role and
`search_path=knowledge`. Run **Knowledge's** migration once with both Knowledge
and Indexer role names. Indexer neither ships nor runs a migration binary.
Fresh roles need SELECT on documents/files/chunks and UPDATE on files/chunks;
they need no artifact, upload, registry, runtime or DDL database privileges.

Each sweep retains the existing three-second cadence, bounded backfill batches
and at most ten pending IDs. A file is processed sequentially. Its owner package
holds one connection's advisory lock and uses another for queries: the pool
must have at least two connections (default three). Multiple processes may run
against this domain; a pending ID is only a hint, and lock acquisition plus
authoritative state re-read occurs inside the owner protocol.

LightRAG's unknown effects require operator reconciliation; a persisted track
ID is polled without uploading again. Milvus retries retain chunk IDs. The
configuration saved with each file is used even if Registry's latest resource
has changed. A provider or database error does not terminate the scheduler;
recovery continues in a later sweep. Cancellation stops further files.

## Compatibility and evidence

Both the Indexer commit and the actual Knowledge module version/checksum belong
in release provenance. The wire/data contract is in
`deploy/release-contract.json`; old imports must remain recoverable across
independent Knowledge/Indexer upgrades and rollback. No reverse migrations.

Unit tests exercise scheduling failure/cancellation. Knowledge's owner tests
exercise the shared protocol with real PostgreSQL, both restricted roles,
process termination during an external upload, unknown-effect handling and
stable Milvus retry IDs; external providers in these fixtures are synthetic.
Neither fixture success nor compilation proves a real configured RAG backend
or all production HA/scale targets. Windows release testing is still pending.
