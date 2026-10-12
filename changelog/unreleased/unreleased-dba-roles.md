# Unreleased: Database roles

- Date: 2026-10-12
- Bookmark: samw/ai/ux-fixes
- Components: hephaestus
- Changelog: Three database roles optimize how the repo uses its databases: `dba/sqlite`, `dba/postgres`, and `dba/meilisearch` check keys, indexes, query patterns, data consistency, and migrations, with evidence before and after.

## Summary

A human asked for a role for DBA tasks, one for each type of database: SQLite, PostgreSQL, and
Meilisearch for now.

## Changes

- `hephaestus/roles/dba/sqlite/role.md`: `STRICT` tables, keys and `WITHOUT ROWID`, indexes that
  `EXPLAIN QUERY PLAN` shows, WAL and the other `PRAGMA` settings, transactions, and
  `integrity_check` and `foreign_key_check`.
- `hephaestus/roles/dba/postgres/role.md`: keys, constraints, and types; indexes (also for foreign
  keys, partial, covering, and GIN or GiST) from `EXPLAIN (ANALYZE, BUFFERS)`; N+1 queries and
  `OFFSET` pagination; migrations with no long lock; vacuum, bloat, and connections.
- `hephaestus/roles/dba/meilisearch/role.md`: the primary key; searchable, filterable, sortable, and
  displayed attributes; ranking and typo settings; limits and document size; the sync of the index
  with the source data; and its task queue.
- `hephaestus/roles/dba/README.md`: the rules of every DBA: never production, evidence before and
  after, consistency first, safe migrations, parameters and never strings, and one change for each
  finding.
- The implementer and the performance engineer name the DBA roles in `Works with`.

## Issues

- Created [[dba-roles]].
