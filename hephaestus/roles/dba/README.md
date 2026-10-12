# Database roles

The database roles (DBA) optimize how the repo uses a database: its schema, keys, indexes, queries,
data consistency, and migrations. Each is an AI role for one database.

| Role | Database | What it checks most |
|---|---|---|
| `dba/sqlite` | SQLite | `STRICT` tables and types, keys and `WITHOUT ROWID`, indexes that query plans use, `PRAGMA` settings (WAL, foreign keys, `busy_timeout`), transactions and locks, and `integrity_check` |
| `dba/postgres` | PostgreSQL | keys and constraints, indexes (B-tree, partial, covering, GIN/GiST) and the foreign keys that need one, query plans, N+1 queries, isolation and locks of migrations, vacuum and bloat, and connection use |
| `dba/meilisearch` | Meilisearch | the primary key, searchable, filterable, and sortable attributes, ranking rules, typo and pagination settings, document size, how the index stays in sync with the source data, and its task queue |

The role files: [[roles/dba/sqlite/role]], [[roles/dba/postgres/role]], and
[[roles/dba/meilisearch/role]].

## Rules for every DBA

- **Never production.** Work on a scratch database: a copy of the schema from the migrations of the
  repo, and test or generated data. Never connect to a database that a person or a service uses,
  unless a human gives read-only access for one task.
- **Evidence before and after.** Show each finding with a query plan, a measurement, or a query that
  finds the bad data, and show each fix with the same evidence after it.
- **Consistency first.** Keys, constraints, and types keep the data correct. A change that makes a
  query faster but lets bad data in is worse.
- **Safe migrations.** A migration can run forward on the data that exists, does not lock a large
  table for long, and has a way back, or says why it has none.
- **Parameters, never strings.** A query that puts a value into its text, not a parameter, is a
  finding for the DBA and for the security researcher.
- **One change for each finding,** on its own bookmark, with tests. A fix of the schema needs a
  migration, never an edit of an old one.

To add a database, add a directory with a `role.md`, and a row to the table.
