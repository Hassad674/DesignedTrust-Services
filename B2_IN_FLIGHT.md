# B.2 in flight

Audit logs cold storage to R2 — extends the B.1 retention sweep so that
`audit_logs_archive` rows older than the cold-tier cutoff are dumped to
R2 as gzipped JSONL files before being hard-deleted from Postgres.

Scope locked, agent dispatched 2026-05-10. This file is a placeholder
to satisfy the robustness rule "first commit ≤ 3 tool uses". It will be
deleted in a follow-up commit.
