# qwen3:4b as a sisyphus task manager

## Facts

- Model: qwen3:4b, 4.0B parameters, Q4_K_M. Ollama 0.40.2 on a Windows workstation GPU with 16 GB VRAM.
- Loaded size: 3,873,366,343 bytes (3.87 GB). All of it is in VRAM.
- Settings: num_ctx 8192, temperature 0.2, max_turns 8, 2 repeats per task, seed 7 and 8.
- Suite: 14 tasks. t01-t05 read, t06-t12 write, t13-t14 multi-step. Each run starts from a fresh repo with 6 seed issues.
- All 13 sisyphus-mcp tools are given to the model ("tools all").
- Configs:
  - A: think off, prompt "basic", schema raw (the optional `dir` argument is visible).
  - B: think off, prompt "detailed", schema nodir (`dir` is hidden and removed).
  - C: think on, prompt "detailed", schema nodir.
- The oracle run (`harness.py --oracle`) passes all 14 checkers.
- Harness bug in A and B: the checker read the issue state from the parent directory. A nested name such as `login-api/add-rate-limiting` failed "state open" wrongly. C used the fixed checker. I re-scored A and B with this rule: a failed run passes if its only failed check is "state open" and the create call returned a path that starts with `issues/open/`.
  - Corrected: A t06 rep0. The call was `sisyphus_new {"name": "login-api/add-rate-limiting", "priority": "high", "effort": "small", "tags": "auth,feature"}`. The output was `issues/open/login-api/add-rate-limiting.md`.
  - No run in B needed a correction.
- "Think off" does not stop the reasoning. In A, 26 of 28 final answers contain a reasoning text that ends with `</think>`. In B, 27 of 28 do. The model writes its reasoning into the visible answer. In C, the final answers are short and clean (60 characters on average, against 1,659 in A and 2,208 in B).
- The model never answered without a tool call (no_tool_runs = 0 in all configs). It never called an unknown tool.

## Results

| | A | B | C |
|---|---|---|---|
| Corrected pass rate | **23/28 (82.1%)** | **24/28 (85.7%)** | **24/28 (85.7%)** |
| Raw pass rate | 22/28 (78.6%) | 24/28 (85.7%) | 24/28 (85.7%) |
| Read (corrected) | 10/10 | 9/10 | 10/10 |
| Write (corrected) | 11/14 | 11/14 | 11/14 |
| Multi (corrected) | 2/4 | 4/4 | 3/4 |
| Generation speed | 186.9 tok/s | 187.0 tok/s | 182.9 tok/s |
| Prompt speed | 29,082 tok/s | 34,784 tok/s | 47,415 tok/s |
| Mean wall time per task | 14.4 s | 16.7 s | 14.9 s |
| Median wall time per task | 7.4 s | 9.6 s | 12.1 s |
| Mean output tokens per task | 2,542 | 3,000 | 2,621 |
| Mean wall time, passed / failed runs | 9.5 s / 32.5 s | 12.6 s / 41.4 s | 10.8 s / 39.4 s |
| Tool calls (per run) | 64 (2.29) | 61 (2.18) | 45 (1.61) |
| Tool-call errors | 22 | 12 | 8 |
| Runs with no tool call | 0 | 0 | 0 |
| Runs that hit max_turns | 2 | 1 | 0 |
| Loaded VRAM | 3.87 GB | 3.87 GB | 3.87 GB |

Tool-call errors by type:

| Error | A | B | C |
|---|---|---|---|
| `No issue named '...'` (guessed name) | 10 | 7 | 6 |
| `unexpected additional properties` | 4 | 3 | 0 |
| `chdir <dir>: no such file or directory` (junk `dir`) | 5 | hidden | hidden |
| `missing properties: ["state"]` (update without state) | 2 | 2 | 1 |
| Wrong type (`text` given a string) | 1 | 0 | 0 |
| `Issue '...' already exists` | 0 | 0 | 1 |

Notes:

- A failed run costs about 3 to 4 times the time of a passed run. Most of the time goes to failed retries, not to work.
- The prompt per turn is 3.2k to 6.3k tokens on average. The tool list alone is about 3.2k tokens. The 8-turn runs (t08) average 6.2k to 6.3k tokens per turn, so the last turns can be near or over num_ctx 8192. Long failed runs can lose early context.

## Tasks

Passes per task, out of 2 (corrected):

| Task | Kind | A | B | C | Note |
|---|---|---|---|---|---|
| t01 high/critical open issues | read | 2 | 2 | 2 | |
| t02 timeout issue | read | 2 | 2 | 2 | |
| t03 docs owner and bookmark | read | 2 | **1** | 2 | B rep0 said the bookmark is "not listed" |
| t04 sub-issues of auth-epic | read | 2 | 2 | 2 | |
| t05 payment issue (none) | read | 2 | 2 | 2 | |
| t06 create rate-limit issue | write | 2 (raw 1) | 2 | 2 | A rep0 corrected |
| t07 start fix-login-timeout | write | 2 | 2 | 2 | |
| t08 close docs theme issue | write | **0** | **0** | **0** | never finds `update-docs-site` |
| t09 abandon password reset | write | 1 | 1 | 2 | |
| t10 set parent | write | 2 | 2 | 2 | |
| t11 add dependency | write | 2 | 2 | 2 | |
| t12 lower priority | write | 2 | 2 | **1** | C rep1 called `sisyphus_new` |
| t13 create with parent and dependency | multi | 2 | 2 | 2 | B rep0 and C both reps used the nested name `auth-epic/audit-session-tokens` |
| t14 find critical, start it, give title | multi | **0** | 2 | 1 | |

- Easy and stable: t01, t02, t04, t05, t07, t10, t11, t13. The user names the issue, or one filter or one word finds it.
- Hard: t08 fails in all 6 runs. The user says "docs site theme update". The issue is `update-docs-site` with title "Update the docs site theme". No search phrase that the model tried matches it.
- t13 passes, but the model did not do what the user asked. The user asked for the name `audit-session-tokens`. In 3 of 6 runs the model made `auth-epic/audit-session-tokens`, because "put it under auth-epic" looks like a directory to it. The checker matches the file stem, so it passes.

## Failure modes

Failed runs (corrected): A 5, B 4, C 4. One run can show more than one mode. The table counts runs that fail and show the mode.

| Mode | A | B | C |
|---|---|---|---|
| 1. Guessed issue name (not taken from a tool result) | 4 | 2 | 2 |
| ...of which the name is nested like `web/auth/fix-login-bug` | 4 | 1 | 1 |
| 2. Search with a phrase that returns nothing | 2 | 3 | 2 |
| 3. `sisyphus_slug` used to "find" an existing issue | 2 | 2 | 2 |
| 4. Wrong final answer ("the issue does not exist", "bookmark not listed") | 1 | 3 | 3 |
| 5. Skips the write step, then answers only the question part | 2 | 0 | 1 |
| 6. Unknown argument (`query` on list/show, `description` on new) | 2 | 2 | 0 |
| 7. Junk `dir` argument | 2 | n/a | n/a |
| 8. Loop until max_turns | 2 | 1 | 0 |
| 9. Wrong tool | 1 | 0 | 1 |
| 10. Repeats the same call with no change | 1 | 1 | 0 |
| 11. Stray mutation | 1 | 0 | 0 |
| 12. Hallucinated answer without tools | 0 | 0 | 0 |

Every failed run per config:

- A: t08 rep0 (1, 2, 3, 4), t08 rep1 (1, 2, 3, 6, 7, 8), t09 rep1 (1, 6, 7, 8, 9, 11), t14 rep0 (1, 5), t14 rep1 (5, 10).
- B: t03 rep0 (4, 6, 10), t08 rep0 (1, 2, 3, 6, 8), t08 rep1 (1, 2, 3, 4), t09 rep1 (2, 4).
- C: t08 rep0 (1, 2, 3, 4), t08 rep1 (1, 2, 3, 4), t12 rep1 (4, 9), t14 rep1 (5).

Examples of the top modes:

1. Guessed nested name. The model copies the shape of the example `web/auth/fix-login-bug` from the tool descriptions:
   - A t09 rep1: `sisyphus_update {"name": "web/auth/password-reset", "state": "closed", "resolution": "abandoned"}` returns `No issue named 'web/auth/password-reset' ...`.
   - A t14 rep0: `sisyphus_list {"priority": "critical"}` returns `migrate-database`. Then the model calls `sisyphus_update {"state": "in-progress", "owner": "carol", "bookmark": "ai/db", "name": "ai/db/fix-critical-bug"}`. It ignores the name it just got and builds one from the bookmark and the nested pattern.
   - Other guesses: `docs/theme/update`, `docs/theme-update` (A, B, C), `docs/site`, `docs-theme-update`, `password-reset`.
2. Search phrase returns nothing. Search matches the whole query as one phrase:
   - A, B, C t08: `sisyphus_search {"query": "docs site theme update"}` returns `[]`. Title is "Update the docs site theme". All words are there, but not in that order. Also `"theme update"` and `"docs theme update"` return `[]`.
   - B t09 rep1: `sisyphus_search {"query": "password reset feature"}` returns `[]` two times. The model then says "No password reset feature issue found." The title is "Add password reset by email".
3. Slug used as a lookup. After the empty search, every t08 run does this:
   - `sisyphus_slug {"text": "docs site theme update"}` returns `docs-site-theme-update`. Then `sisyphus_update {"name": "docs-site-theme-update", "state": "closed", ...}` returns `No issue named ...`. The slug tool makes a name for a new issue. The model uses it to guess the name of an existing issue.
4. Unknown argument. `sisyphus_list {"query": "password reset"}` returns `unexpected additional properties ["query"]` (A t05, t08, t09; B t08). The model expects `query` on list because search has it.
5. Skips the write. A t14 rep1 and C t14 rep1 call `sisyphus_list {"priority": "critical", "state": "open"}`, see `migrate-database`, and then reply "Migrate to Postgres 17". They never call `sisyphus_update`.

Other details:

- B t03 rep0: `sisyphus_search` returns `owner` but not `bookmark`. The model then sends `sisyphus_show {"query": "docs site"}` (error), repeats the same search 4 times, and says the bookmark "is not explicitly listed in the tool response".
- C t12 rep1: the model wants to change the priority. It calls `sisyphus_new {"name": "migrate-database", ..., "priority": "critical", ...}` with the template text as `context`. The tool returns `Issue 'migrate-database' already exists`. The model replies "Issue 'migrate-database' already exists and is open." Nothing changes.
- A t09 rep1 is the worst run: junk `dir` values `"current"` and `"web"`, two calls to `sisyphus_init`, an unknown `description` argument, and then `sisyphus_new {"name": "password-reset"}`. This creates a stray issue.
- `sisyphus_update` without `state` happens in t12 in 5 of 6 runs. The model always recovers on the next call, but it costs one turn (about 5 to 15 s).

## Effect of settings

### A to B: detailed prompt and hidden `dir`

- Corrected pass rate: 23 to 24 of 28. Raw: 22 to 24.
- Fixed: t14 (0/2 to 2/2). The prompt line "To start work: sisyphus_update with state in-progress" helps the model do the write step.
- Fixed only in the raw score: t06 rep0 (A used a nested name; B did not).
- Broke: t03 rep0 (bookmark missing from the answer). This looks like noise plus mode 10, not a prompt effect.
- No change: t08 is still 0/2. t09 is still 1/2, but the failure changed. A failed by junk `dir` and a stray issue. B failed by an empty search and a "no such issue" answer.
- Hiding `dir` removed all 5 `chdir` errors and the `sisyphus_init` calls.
- Tool-call errors: 22 to 12. Runs at max_turns: 2 to 1.
- Cost: mean wall 14.4 to 16.7 s (+16%). Mean output tokens 2,542 to 3,000 (+18%). Most of the increase is in t14 (B rep1 took 70 s and 11,961 tokens because it repeated the update 3 times).

### B to C: thinking on

- Corrected pass rate: 24 to 24 of 28.
- Fixed: t03 rep0 and t09 rep1. C found `add-password-reset` with the query "password reset" in both t09 runs.
- Broke: t12 rep1 (wrong tool `sisyphus_new`) and t14 rep1 (skipped the write).
- Fewer calls and errors: 61 to 45 calls, 12 to 8 errors, no unknown arguments, no max_turns.
- Cost: none. Mean wall 16.7 to 14.9 s. Mean output tokens 3,000 to 2,621. The reason: with think off, qwen3:4b still reasons in the visible answer. With think on, the reasoning moves to the thinking field and the answer is short.
- Read tasks are faster with C (6.1 s against 8.9 s). Multi tasks are faster than B (25.9 s against 32.8 s).
- The final answers in C are clean and short. In A and B, the user sees a long reasoning text and a `</think>` tag before the answer.

## Ideas for sisyphus-mcp

1. **Match search words, not one phrase.** Evidence: t08 failed 6/6 and B t09 rep1 failed, all after `sisyphus_search` returned `[]` for queries whose words are all in the title ("docs site theme update", "password reset feature"). Match an issue when all query words (or most of them) are in the name, title, or body, in any order. Also match the issue name.
2. **Suggest close names when a name is not found.** Evidence: 23 `No issue named` errors (10, 7, 6). Change the error to, for example: `No issue named 'docs-site-theme-update'. Similar issues: update-docs-site (Update the docs site theme). Use sisyphus_search or sisyphus_list to find names.` Rank by shared words between the guess and the names and titles. This one change would likely fix t08 and the guessed-name part of t09 and t14.
3. **Remove the nested example from tool descriptions, or explain it.** Evidence: the guesses `web/auth/password-reset`, `ai/db/fix-critical-bug`, `docs/theme/update`, `docs/theme-update`, `docs/site`, and the created names `login-api/add-rate-limiting` and `auth-epic/audit-session-tokens` (3 of the 4 t13 runs in B and C). In the `name` field of show, update, graph, parent, depends_on, and remote, write: "The issue name exactly as sisyphus_list or sisyphus_search returns it, for example fix-login-bug. Do not make up a name." In `sisyphus_new`, say that a directory prefix is optional and is not the parent: "Use sisyphus_parent or the parent field for sub-issues, not a directory."
4. **Say what `sisyphus_slug` is for.** Evidence: every t08 run calls slug and then uses the result as an existing issue name. Start the description with: "Only for naming a NEW issue before sisyphus_new. It does not find existing issues; use sisyphus_search for that."
5. **Make `state` optional in `sisyphus_update`.** Evidence: 5 `missing properties: ["state"]` errors, 5 of 6 t12 runs, each costs a turn. Keep the current state when `state` is empty. C t12 rep1 even switched to `sisyphus_new` to change a priority. Also add "Use this to change priority, effort, or tags" to the description.
6. **Accept `query` on `sisyphus_list`, or merge list into search.** Evidence: 6 `unexpected additional properties ["query"]` errors on list/show. If list stays strict, change the error text to: `sisyphus_list has no "query". Use sisyphus_search for text.` The current schema error does not tell the model which tool to use.
7. **Return `bookmark` (and other set fields) in list and search JSON.** Evidence: B t03 rep0. Search returned `owner` but no `bookmark`, and the model said the bookmark was not listed.
8. **Hide or drop `dir` for agents.** Evidence: A had 5 `chdir` errors with `dir` values `"docs"`, `"current"`, and `"web"`, and two `sisyphus_init` calls. Config B/C (no `dir`) had none. Remove `dir` from the schema, or add a server flag that hides it.
9. **Shorten the tool list.** Evidence: the tools take about 3.2k of the 8,192 context tokens, and `sisyphus_graph` alone has a long description. The 8-turn runs reach about 6.3k tokens per turn on average. Offer a "core" tool set (new, update, show, list, search, parent, depends_on) and a short graph description for small models. `sisyphus_init` should not be in the default set for an agent in a repo that is already set up.
10. **Return a short confirmation on writes.** Evidence: the models read `issues/open/migrate-database.md` from `sisyphus_parent` and `sisyphus_depends_on` as "the file was created" and spend reasoning on it. A result such as `parent of migrate-database is now auth-epic (issues/open/migrate-database.md)` is clearer.

## Verdict

- qwen3:4b is usable as a fast sisyphus task manager for direct tasks, where the user gives the issue name or a single word or filter finds it. These tasks pass in almost every run and take 3 to 10 s.
- It is not reliable when it must find an issue from a loose description. It does not use `sisyphus_list` to browse. It guesses names, often in the nested `dir/name` shape from the tool descriptions. t08 failed in all 6 runs. Fix ideas 1 to 4 in sisyphus-mcp first, then test again.
- Do not trust it alone on multi-step write tasks. In 3 of 6 t14 runs (A both, C one) it found the issue but did not start it. A user must check the result.
- Best config: **C** (think on, detailed prompt, no `dir`). It ties B on pass rate (24/28), but it is faster (14.9 s against 16.7 s), uses fewer tokens and fewer tool calls, has no unknown arguments and no max_turns runs, and gives clean short answers. "Think off" gives no speed gain with qwen3:4b on Ollama 0.40.2, because the model still reasons in the visible answer.
- Speed: about 185 tok/s generation and 3.87 GB VRAM. It leaves most of the 16 GB GPU free.
