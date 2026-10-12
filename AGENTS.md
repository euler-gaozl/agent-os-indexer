# Indexer ownership

Read README.md. This repository owns the Indexer process and sequential schedule.
Knowledge owns its SQL and migrations; consume the narrow `indexing` package at
the fixed module version with committed vendor. Never add copied SQL, migrations,
Knowledge HTTP routes, sibling directory replacements or application credentials.
Use Go, bounded resources, an application database role, and GOWORK=off builds.
Changes to transaction semantics belong in the Knowledge owner and its recovery
tests. No real API keys, database credentials or private documents in Git/logs.
