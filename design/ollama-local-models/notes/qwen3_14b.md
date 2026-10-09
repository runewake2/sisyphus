# qwen3:14b as a sisyphus task manager

## Facts

- Model: qwen3:14b, 14.8B parameters, Q4_K_M, family qwen3. Capabilities: completion, tools, thinking. Ollama 0.40.2 on a Windows workstation GPU with 16 GB VRAM.
- Loaded size: 10,468,448,009 bytes (10.47 GB). All of it is in VRAM. This is the largest model in the test (qwen3:8b 6.30 GB, qwen3.5:9b 5.83 GB).
- Settings: num_ctx 8192, temperature 0.2, max_turns 8, 2 repeats per task, seed 7 and 8.
- Suite: 14 tasks. t01-t05 read, t06-t12 write, t13-t14 multi-step. Each run starts from a fresh repo with 6 seed issues.
- All 13 sisyphus-mcp tools are given to the model ("tools all").
- Configs:
  - A: think off, prompt "basic", schema raw (the optional `dir` argument is visible).
  - B: think off, prompt "detailed", schema nodir (`dir` is hidden and removed).
  - C: think on, prompt "detailed", schema nodir.
- **Config C was merged from two sessions.** The host Ollama crashed during repeat 1. The 8 runs C rep1 t07-t14 failed with `HTTP Error 500` (0 calls, 0 s). These 8 runs were run again with the same seed (8) and settings, and merged. They have `"rerun": true`. I checked the merge: the 20 other C runs are identical to the runs in `rerun/original-with-server-errors.json` (same calls, same final answers). Generation speed is the same in both sessions (76.0 against 77.8 tok/s). I ignore the HTTPError runs.
- Because of the crash, the C file has no `loaded_bytes`, and its header `prompt_tok_per_s` (41,478) is copied from the pre-merge file. The prompt speed in the table below is computed from all 28 merged runs.
- The tested `sisyphus-mcp` binary (`bin/sisyphus-mcp`, built 10:26) comes from the `sisyphus-ollama-model-spike` workspace. Its `name` descriptions contain the nested example `web/auth/fix-login-bug` (17 times in the binary; `cmd/sisyphus-mcp/tools.go` line 93 in that workspace). The main checkout at `53f4d7f` does not have this example (its `newArgs.Name` says only "for example fix-login-bug").
- The `dir` description is the same in the binary and in both checkouts: "The repo's root directory, or a directory below it. Defaults to sisyphus-mcp's own working directory." This sentence is important for config A (see "Failure modes").
- **qwen3:14b obeys "think off".** 0 of 56 think-off final answers contain `<think>`, `</think>`, or reasoning text. The answers are short: 185 characters on average in A (most are "check the directory path" messages), 111 in B. C answers are 166 characters on average.
- The model never answered without a tool call (no_tool_runs = 0 in all configs). It never called an unknown tool. No run hit max_turns (max 7 turns, A t12).
- **3 C final answers are empty** (C t03 rep0 and rep1, C t14 rep0). No other model had an empty final answer. The harness does not save the thinking text, so I cannot see why.
- The harness does not save the thinking text. For C, the thinking size is only visible as output tokens. An estimate of the thinking part is "C tokens minus B tokens" for the same task.

## Results

| | A | B | C |
|---|---|---|---|
| Pass rate | **11/28 (39.3%)** | **24/28 (85.7%)** | **21/28 (75.0%)** |
| Read | 7/10 | 8/10 | 8/10 |
| Write | 4/14 | 13/14 | 12/14 |
| Multi | 0/4 | 3/4 | 1/4 |
| Mean check score | 0.645 | 0.911 | 0.864 |
| Generation speed | 77.8 tok/s | 75.1 tok/s | 77.3 tok/s |
| Prompt speed (cached prefix) | 23,299 tok/s | 56,457 tok/s | 36,705 tok/s |
| Mean wall time per task | 2.71 s | 1.37 s | 8.98 s |
| Median wall time per task | 2.41 s | 1.39 s | 7.85 s |
| Max wall time per task | 10.57 s (A t01 rep0, first run) | 2.81 s | 18.73 s |
| Mean output tokens per task | 157 | 84 | 669 |
| Median output tokens per task | 154 | 87 | 572 |
| Mean wall time, passed / failed runs | 1.71 s / 3.36 s | 1.31 s / 1.75 s | 8.35 s / 10.89 s |
| Tool calls (per run) | 76 (2.71) | 44 (1.57) | 43 (1.54) |
| Mean turns per run | 3.71 | 2.57 | 2.43 |
| Tool-call errors | **63** | 7 | 6 |
| Runs with an error / of which passed | 22 / 5 | 6 / 3 | 5 / 4 |
| Runs with no tool call | 0 | 0 | 0 |
| Runs that hit max_turns | 0 | 0 | 0 |
| Empty final answers | 0 | 0 | 3 |
| Final answers with leaked reasoning | 0 | 0 | 0 |
| Loaded VRAM | 10.47 GB | 10.47 GB | not recorded (crash); same model |

Tool-call errors by type:

| Error | A | B | C |
|---|---|---|---|
| `chdir ...: no such file or directory` (junk `dir`) | **57** | hidden | hidden |
| `Cannot find the repo root` (`"dir": "/home"`) | 2 | hidden | hidden |
| `No issue named '...'` (guessed name) | 2 | 7 | 2 |
| `missing properties: ["state"]` (update without state) | 2 | 0 | 2 |
| `unexpected additional properties` | 0 | 0 | 2 |
| Total | 63 | 7 | 6 |

Per kind, mean wall time and output tokens:

| Kind | A | B | C |
|---|---|---|---|
| Read | 2.34 s, 90 tok | 1.20 s, 69 tok | 6.90 s, 516 tok |
| Write | 3.03 s, 201 tok | 1.47 s, 93 tok | 9.63 s, 713 tok |
| Multi | 2.53 s, 168 tok | 1.45 s, 92 tok | 11.90 s, 900 tok |

Notes:

- Generation is slow per token: 75-78 tok/s. qwen3:8b makes 119-123 tok/s and qwen3.5:9b 114-115 tok/s on the same GPU. qwen3:14b is about 35% slower per token.
- B is still fast (1.4 s), because a B run writes only 84 tokens.
- A is 2 times slower than B, and not because of thinking. A makes 2.71 calls per run, against 1.57 in B. Most of the extra calls are retries with a new junk `dir`.
- The prompt per turn is 3.3k to 3.8k tokens. The context is never a problem.

### Output tokens in C (thinking included)

| | qwen3:14b C | qwen3:8b C | qwen3.5:9b C | qwen3:4b C |
|---|---|---|---|---|
| Min | 251 | 360 | 131 | 565 |
| Median | 572 | 677 | 305 | 2,066 |
| p90 | 1,289 | 1,497 | 554 | 5,856 |
| Max | 1,385 (t08 rep1) | 16,490 (t12 rep1) | 564 | 8,944 |
| Mean | 669 | 1,328 | 309 | 2,621 |
| Total for 28 runs | 18,740 | 37,173 | 8,639 | 73,394 |
| Estimated thinking per run (C minus B) | 585 | 1,238 | 197 | n/a (4b leaks reasoning with think off) |
| Mean / max wall time | 9.0 s / 18.7 s | 11.3 s / 140.7 s | 3.3 s / 5.7 s | 14.9 s / 51.7 s |

- qwen3:14b thinks about half as much as qwen3:8b, and about 3 times as much as qwen3.5:9b.
- There is no runaway thinking. The max is 1,385 tokens. The longest runs are the failed t08 runs (1,311 and 1,385 tokens). The model thinks more when it cannot find the issue, but then it stops and asks the user.
- The extra thinking does not buy extra checks. C makes fewer calls than B (43 against 44). qwen3.5:9b C, for comparison, made more calls than its B (49 against 47).

Calls and output tokens per task in C (2 runs each):

| Task | 14b C calls | 14b C tokens | 14b B tokens | 8b C tokens | 3.5:9b C tokens |
|---|---|---|---|---|---|
| t01 | 1, 1 | 818, 806 | 71, 71 | 819, 665 | 193, 193 |
| t02 | 1, 1 | 479, 644 | 42, 42 | 718, 410 | 131, 133 |
| t03 | 1, 1 (fail, fail) | 251, 251 | 85, 85 | 904, 993 | 263, 269 |
| t04 | 1, 1 | 468, 468 | 98, 98 | 595, 449 | 231, 238 |
| t05 | 1, 1 | 460, 515 | 51, 51 | 377, 360 | 387, 554 |
| t06 | 1, 1 | 579, 603 | 95, 95 | 879, 568 | 366, 379 |
| t07 | 1, 1 | 582, 565 | 88, 88 | 848, 803 | 564, 354 |
| t08 | 3, 2 (fail, fail) | 1,311, 1,385 | 178, 172 | 647, 651 | 329, 347 |
| t09 | 3, 2 | 685, 560 | 118, 118 | 1,497, 1,246 | 286, 287 |
| t10 | 1, 1 | 411, 376 | 56, 56 | 591, 597 | 227, 232 |
| t11 | 1, 1 | 364, 432 | 62, 62 | 533, 574 | 238, 203 |
| t12 | 3, 3 | 1,289, 839 | 57, 57 | 1,616, 16,490 | 347, 325 |
| t13 | 1, 1 (fail, fail) | 1,153, 469 | 92, 94 | 688, 595 | 560, 338 |
| t14 | 2, 5 (fail, pass) | 933, 1,044 | 92, 92 | 957, 1,103 | 342, 323 |

- C t03 uses only 251 tokens in each run, the smallest of all C runs, and ends with an empty answer. So the t03 failure is not overthinking. The model stops.

## Tasks

Passes per task, out of 2:

| Task | Kind | A | B | C | Note |
|---|---|---|---|---|---|
| t01 high/critical open issues | read | 1 | 2 | 2 | A rep0: two junk `dir` values, then asks for the path |
| t02 timeout issue | read | 2 | 2 | 2 | |
| t03 docs owner and bookmark | read | **0** | **0** | **0** | A junk `dir`; B "no owner, no bookmark"; C empty answer |
| t04 sub-issues of auth-epic | read | 2 | 2 | 2 | |
| t05 payment issue (none) | read | 2 | 2 | 2 | reads `[]` correctly in all 6 runs |
| t06 create rate-limit issue | write | 2 | 2 | 2 | A recovers with `"dir": "."` |
| t07 start fix-login-timeout | write | **0** | 2 | 2 | A: 3 junk `dir` values per run |
| t08 close docs theme issue | write | **0** | 1 | **0** | phrase search returns nothing |
| t09 abandon password reset | write | 2 | 2 | 2 | |
| t10 set parent | write | **0** | 2 | 2 | A junk `dir` |
| t11 add dependency | write | **0** | 2 | 2 | A junk `dir` |
| t12 lower priority | write | **0** | 2 | 2 | A walks up from `/home/user/repo` to `/home` |
| t13 create with parent and dependency | multi | **0** | 1 | **0** | B and C: parent put in the name; A: parent put in `dir` |
| t14 find critical, start it, give title | multi | **0** | 2 | 1 | C rep0 stops with an empty answer before the update |

Side by side with the best configs of the other local models (passes out of 2):

| Task | 14b A | 14b B | 14b C | 4b C | 8b B | 8b C | 3.5:9b C | ministral B |
|---|---|---|---|---|---|---|---|---|
| t01 | 1 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t02 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t03 | 0 | 0 | 0 | 2 | 0 | 2 | 2 | 0 |
| t04 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t05 | 2 | 2 | 2 | 2 | 0 | 2 | 2 | 2 |
| t06 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t07 | 0 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t08 | 0 | 1 | 0 | 0 | 1 | 0 | 2 | 0 |
| t09 | 2 | 2 | 2 | 2 | 2 | 1 | 2 | 2 |
| t10 | 0 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t11 | 0 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t12 | 0 | 2 | 2 | 1 | 2 | 1 | 2 | 1 |
| t13 | 0 | 1 | 0 | 2 | 2 | 2 | 2 | 0 |
| t14 | 0 | 2 | 1 | 1 | 2 | 2 | 2 | 0 |
| **Total** | **11/28** | **24/28** | **21/28** | 24/28 | 23/28 | 24/28 | **28/28** | 19/28 |
| Mean wall time | 2.7 s | 1.4 s | 9.0 s | 14.9 s | 0.9 s | 11.3 s | 3.3 s | 0.9 s |
| Mean output tokens | 157 | 84 | 669 | 2,621 | 90 | 1,328 | 309 | 78 |
| Generation speed | 78 tok/s | 75 tok/s | 77 tok/s | 183 tok/s | 119 tok/s | 123 tok/s | 115 tok/s | 112 tok/s |
| VRAM | 10.47 GB | 10.47 GB | 10.47 GB | 3.87 GB | 6.30 GB | 6.30 GB | 5.83 GB | 6.34 GB |

Config A of the other models, for comparison: qwen3:4b 22-23/28, qwen3:8b 18/28, qwen3.5:9b 23/28, ministral-3:8b 17/28. qwen3:14b A (11/28) is the lowest score of all configs of all models.

- Easy and stable for qwen3:14b (B and C): t01, t02, t04, t05, t06, t07, t09, t10, t11, t12. All 4 runs pass. The user names the issue, or one filter or one word finds it.
- t05 passes in all 6 runs. qwen3:14b reads an empty result correctly. It never invents issues (unlike qwen3:8b B).
- t12 passes in B and C. B sends `"state": "open"` at once. C forgets `state`, then calls `sisyphus_show` and retries (2 of 2 recover).
- t03 fails in all 6 runs. Only qwen3:14b and ministral-3:8b fail t03 in every config.
- t13 fails in 5 of 6 runs. qwen3:8b passed it in all 6 runs. See "Failure modes".

## Failure modes

Failed runs: A 17, B 4, C 7. One run can show more than one mode. The table counts failed runs that show the mode. The numbering follows the earlier notes. Modes 24-26 are new for qwen3:14b.

| Mode | A | B | C |
|---|---|---|---|
| 1. Guessed issue name (not taken from a tool result) | 4 | 3 | 1 |
| ...of which the name is nested like `theme/update` | 2 | 0 | 0 |
| 2. Search that returns nothing (phrase or name as query) | 0 | 1 | 2 |
| 4. Wrong final answer to a read task | 3 | 2 | 2 |
| 5. Skips the write step | 0 | 0 | 1 |
| 6. Unknown argument | 0 | 0 | 1 |
| 7. **Junk `dir` argument** | **17** | n/a | n/a |
| 8. Loop until max_turns | 0 | 0 | 0 |
| 10. Repeats the same call with no change | 10 | 0 | 0 |
| 11. Stray mutation | 0 | 1 | 0 |
| 12. Answered without tools | 0 | 0 | 0 |
| 13. Gives up and asks the user | 17 | 0 | 2 |
| 14. Invents issue data after an empty tool result | 0 | 0 | 0 |
| 15. `sisyphus_update` without `state` | 2 | 0 | 0 |
| 16. Runaway thinking | 0 | 0 | 0 |
| 17. Reads a missing field as "not set" | 0 | 2 | 0 |
| 19. Parent given as a directory (in the name or in `dir`) on create | 2 (in `dir`) | 1 | 2 |
| 24. **Empty final answer (think on)** | n/a | n/a | 3 |
| 25. **Creates a new issue when the named one is not found** (mode 11 above) | 0 | 1 | 0 |
| 26. **Location words from the prompt put in `dir`** (`docs-site`, `auth-epic`) | 4 | n/a | n/a |

Every failed run per config:

- A: t01 rep0 (7, 13), t03 rep0 (1, 4, 7, 13), t03 rep1 (1, 4, 7, 10, 13), t07 rep0 and rep1 (7, 10, 13), t08 rep0 and rep1 (1 nested, 7, 10, 13, 26), t10 rep0 and rep1 (7, 13), t11 rep0 and rep1 (7, 10, 13), t12 rep0 and rep1 (7, 13, 15), t13 rep0 and rep1 (7, 10, 13, 19, 26), t14 rep0 (7, 10, 13), t14 rep1 (7, 13).
- B: t03 rep0 and rep1 (1, 4, 17), t08 rep0 (1, 2, 11, 25), t13 rep1 (19).
- C: t03 rep0 and rep1 (4, 24), t08 rep0 (1, 2, 6, 13), t08 rep1 (2, 13), t13 rep0 and rep1 (19), t14 rep0 (5, 24).

### Why A is so poor: the `dir` argument

All 17 failed A runs have a `chdir` error. Without the `dir` problem, A would be close to B.

- The model sends `dir` in 70 of 76 calls. qwen3:8b sent it in 28 of 38 calls, ministral-3:8b in 6 of 33, qwen3.5:9b in 0 of 46.
- The only value that works is `"."` (9 calls). The first call of a run uses `"dir": "sisyphus-mcp"` in 15 of 28 runs. This value comes from the `dir` description: "Defaults to **sisyphus-mcp**'s own working directory." The model reads the tool name as a directory name.
- The other values are invented paths: `/Users/bob/Projects/sisyphus-mcp` (4), `/home/username/repo` (4), `/home/user/repo/sisyphus-mcp` (3), `/home/runner/work/sisyphus-mcp/sisyphus-mcp` (2), `/home/carol/sisyphus-mcp` (2), `/path/to/repo` (2), `/path/to/your/repo` (1), `your-repo-directory` (1). The model uses names from the prompt (bob, carol) to build the path.
- Mode 26: in t08 and t13, the model puts a location word from the user prompt in `dir`. t08 "The docs site theme update": `{"dir": "docs-site", "name": "theme/update"}`, then `./docs-site`, `/docs-site`, `docs-site`. t13 "Put it under auth-epic": `{"dir": "auth-epic", "name": "audit-session-tokens", ...}`, then `/auth-epic`, then `auth-epic` again. So "under auth-epic" goes into `dir` in A, and into the name in B and C.
- After a `chdir` error, the model almost never drops `dir`. It tries a new invented path. It recovered with `"dir": "."` in only 3 runs (A t01 rep1, A t06 rep0 and rep1). In 10 runs, it repeated an identical failed call (mode 10).
- A t12 walks up the tree: `/home/user/repo/sisyphus-mcp`, `/home/user/repo`, `/home/user`, `/home`. The last one gives `Cannot find the repo root.` Then it stops.
- Every failed A run ends with a request for the path (mode 13). Examples:
  - A t07 rep0: "It seems there is an issue with the directory path. Please ensure that the directory `/Users/bob/Projects/sisyphus-mcp` exists and is accessible."
  - A t10 rep1: "... navigate to your repository's root directory in the terminal and use the `pwd` command to find the current working directory. Then, use that path in the function call."
- A t06 (both runs) called `sisyphus_init {"dir": "sisyphus-mcp", "force": true}` after the `chdir` error. It failed on the same `chdir`. With a valid `dir`, `--force` can overwrite `issues/TEMPLATE.md`. No damage happened, but this is a risky reflex.
- A is not poor because of the prompt only. The calls are mostly correct apart from `dir`. In t07, t10, t11, t13, and t14, the `name`, `state`, `owner`, `bookmark`, `parent`, and `blockingIssue` values are correct. Only `dir` breaks the call.

### Why t03 fails in all 6 runs

The task: "Who owns the docs site issue, and what bookmark is it on?" The issue is `update-docs-site`, owner alice, bookmark `ai/docs-theme`.

- A (both runs): `sisyphus_show {"dir": "sisyphus-mcp", "name": "docs-site", "text": true}`. `chdir` error. Then asks for the path.
- B (both runs): `sisyphus_show {"name": "docs-site"}` returns `No issue named 'docs-site'`. Then `sisyphus_search {"query": "docs site", "text": true}` finds the issue. The text output of search is `update-docs-site: Update the docs site theme` plus the body. It has **no owner and no bookmark**. Final: "The issue `update-docs-site` is owned by no one (owner is not set), and there is no bookmark associated with it." Both values are wrong. The model reads the missing fields as "not set" (mode 17, as ministral-3:8b).
- C (both runs): `sisyphus_search {"query": "docs site"}` (JSON). The row has `"owner": "alice"` but no `bookmark`. Then the final answer is **empty** (251 tokens in total). The model does not call `sisyphus_show`. qwen3:8b C and qwen3.5:9b C called `sisyphus_show` here and passed.
- So t03 fails for the same sisyphus-mcp reason as before: list and search rows have no `bookmark`, and text search rows have no owner either. The main checkout still has `listRow` and `searchRow` with `Owner` and no `Bookmark` (`cmd/sisyphus/list.go`, `cmd/sisyphus/search.go`).

### Why t08 fails in 5 of 6 runs

The task: "The docs site theme update is done. Close it." The title is "Update the docs site theme".

- A (both runs): `sisyphus_update {"dir": "docs-site", "name": "theme/update", ...}` 4 times with 3 different `dir` values. A nested guess and a junk `dir` together.
- B rep0: `sisyphus_update {"name": "docs-site-theme-update", ...}` returns `No issue named`. Then 3 searches for "docs site theme update" in the states closed, open, and in-progress. All return `[]`. Then the model **creates a new issue**: `sisyphus_new {"name": "docs-site-theme-update", "state": "closed", "resolution": "completed"}`. Final: "The issue `docs-site-theme-update` has been created and closed with resolution "completed"." The real issue stays in-progress. This is a stray write (mode 25).
- B rep1 (pass): `sisyphus_update {"name": "docs-site/theme-update"}` (nested, not found), search (empty string), the same update again, then `sisyphus_list {"text": true}`. The list shows `update-docs-site`. Then the correct update. This is the only t08 pass of qwen3:14b. It is the same "browse with list" move as qwen3:8b B rep1.
- C rep0: `sisyphus_search {"query": "docs site theme update", "state": "open"}` (empty), `sisyphus_search {"name": "docs-site-theme-update"}` (`unexpected additional properties ["name"]`), `sisyphus_search {"query": "docs-site-theme-update"}` (empty). Final: "The issue "docs-site-theme-update" was not found in the open state. Please verify the issue name ... you may need to create it first using `sisyphus_new`." The issue is in-progress, and the model only searched open.
- C rep1: search "docs site theme update" in open,in-progress and then in closed. Both `[]`. Final: "No matching issue found. Please confirm the issue name or check for typos."
- So t08 fails for the same sisyphus-mcp reasons as before: the search is one phrase match (`strings.Contains` on the full query, `cmd/sisyphus/search.go` line 48). "docs site theme update" does not match "Update the docs site theme". And the guessed names (`theme/update`, `docs-site/theme-update`) follow the nested example.
- C searches more and thinks more (1,311 and 1,385 tokens), but never tries a shorter query or `sisyphus_list`. qwen3.5:9b C dropped to "theme" or "docs" after one empty search and passed both runs.

### Why multi-step fails in C (1 of 4)

- **t13 (0 of 2 in C, 1 of 2 in B): the parent goes into the name.** All 4 B and C runs create `auth-epic/audit-session-tokens`:
  - C rep0: `sisyphus_new {"name": "auth-epic/audit-session-tokens", "title": "Audit session tokens", "priority": "medium", "dependsOn": "fix-login-timeout"}`. No `parent`. Final: "The issue `audit-session-tokens` has been created under the `auth-epic` directory ... No further action is needed for the dependency". The model says "directory", so it knows it made a directory, not a parent link.
  - C rep1: the same call. Final: "The issue `auth-epic/audit-session-tokens` has been created ...".
  - B rep1: the same call. Fail.
  - B rep0: the same nested name, **plus** `"parent": "auth-epic"`. Pass (the checker reads the state directory). This is the only pass.
  - The dependency is correct in all 4 runs (`dependsOn` in the `new` call). Only the parent is lost.
  - This is the same failure as ministral-3:8b (mode 19). It comes from the nested example in the tested binary: "optionally below kebab-case directories, for example fix-login-bug or web/auth/fix-login-bug". "Put it under auth-epic" matches "below directories" better than the `parent` description ("The parent issue, if this is a sub-issue"). qwen3:8b never did this. qwen3:14b reads the description more literally.
  - Thinking does not help: C rep0 used 1,153 tokens for this one call.
- **t14 (1 of 2 in C, 2 of 2 in B): the model stops before the write.**
  - C rep0: `sisyphus_list {"priority": "critical", "state": "open,in-progress", "text": true}` finds `migrate-database`. `sisyphus_show {"name": "migrate-database"}` returns the issue. Then the final answer is **empty**. No update, no title. 933 tokens.
  - C rep1 (pass): after the list, the model calls `sisyphus_update {"name": "critical-issue-name", ...}`. This is a placeholder name, not a real name. It gets `No issue named`, calls `sisyphus_show {"name": "critical-issue-name"}` (also not found), then does the correct update and a check with `sisyphus_show`. 5 calls.
  - B (both runs): list, update, answer. 2 calls, 92 tokens.

### Other observations (no failed run)

- **Nested names on create:** A t06 (both runs) creates `auth/login-rate-limit`. B and C t06 create `add-rate-limiting-to-login-api` (no nesting). In t13, 4 of 4 B and C runs nest (see above).
- **C t09 rep1 changes the tags:** `sisyphus_update {"name": "add-password-reset", "state": "closed", "resolution": "abandoned", "tags": "abandoned,feature"}`. The tag `auth` is lost. The checker does not test tags on the target issue, so the run passes. This is a small stray change.
- **C t09 rep0 invents an argument:** `sisyphus_update {"current_state": "open", ...}`. This comes from the prompt line "pass the current state". The schema error is clear, and the model retries without it.
- **C t10 wording:** "The sub-issue `migrate-database` has been created under `auth-epic`" (both runs). The parent link is correct, but the issue was not created. Not checked.
- C t05 and B t05 offer to create an issue ("Would you like to create a new issue for this?"). This is fine.

## Effect of settings

### A to B: detailed prompt and hidden `dir`

- Pass rate: 11 to 24 of 28. This is the largest A-to-B gain of all models (qwen3:8b +5, qwen3.5:9b +1, ministral-3:8b +2).
- Fixed: t01 rep0, t07, t10, t11, t12, t14 (11 runs). All of them failed only because of `dir` in A. In B, the same calls without `dir` pass at once (1 call per run, except t14 with 2).
- Fixed, partly: t08 (0/2 to 1/2), t13 (0/2 to 1/2).
- Broke: nothing at task level. t03 stays at 0/2, with a new failure (owner and bookmark "not set").
- New risk: B t08 rep0 creates a new closed issue when it cannot find the target. A could not do this, because every A write failed on `dir`.
- Tool calls: 76 to 44. Tool-call errors: 63 to 7 (all 7 are `No issue named`). Runs with an error that then passed: 5 of 22 to 3 of 6.
- Cost: none. B is faster. Mean wall time 2.71 to 1.37 s. Mean output tokens 157 to 84.
- Most of the gain comes from hiding `dir`. The prompt helps in t12 (B sends `"state": "open"` at once, A forgot `state`) and in t09 (both pass). But the A t12 runs also failed on `dir` after the state retry.

### B to C: thinking on

- Pass rate: 24 to 21 of 28. qwen3:14b is the only model where thinking lowers the score (qwen3:8b +1, qwen3.5:9b +4).
- Broke (3 runs):
  - t14 rep0: the model stops with an empty answer after `sisyphus_show`. It does not do the update (modes 5 and 24).
  - t13 rep0: the same nested name as B, but without the extra `"parent": "auth-epic"` that saved B rep0. B rep1 failed the same way, so this is partly noise with 2 repeats.
  - t08 rep1: B rep1 passed with `sisyphus_list`. C never lists. It searches the same phrase in other states and asks the user.
- Fixed: nothing. t03 still fails, now with an empty answer.
- So the cause is **not** overthinking (no runaway, max 1,385 tokens) and **not** a wrong dependency (`dependsOn` is correct in all t13 runs). The causes are:
  1. **Stops early.** 3 empty final answers (t03 both runs, t14 rep0). In t14 rep0 the model has all it needs and makes no write. In t03 it has the owner and makes no `sisyphus_show` call.
  2. **Asks the user instead of browsing.** t08 both runs. No `sisyphus_list`, no shorter query.
  3. **Wrong parent.** t13: the parent goes into the name as a directory. This is the same in B, so thinking does not fix it.
- C does fewer calls than B (43 against 44), although it thinks 585 tokens more per run. For qwen3.5:9b, thinking led to an extra `sisyphus_show` before a write or after a search. For qwen3:14b, it does not.
- C fixes the missing-state error by itself: t12 forgets `state` in both runs, then calls `sisyphus_show` and retries with `"state": "open"`. B did not forget `state`.
- Cost: large. Mean wall time 1.37 to 8.98 s (6.6 times). Median 1.39 to 7.85 s. Mean output tokens 84 to 669 (8 times).

## Ideas for sisyphus-mcp

The ideas are in order of evidence for qwen3:14b. For each idea, the last line says if it confirms, weakens, or is new compared with the earlier notes.

1. **Hide or drop `dir` for agents, and fix its description.** Evidence: 57 `chdir` errors and 2 `Cannot find the repo root` errors in A. All 17 failed A runs have a `chdir` error. Hiding `dir` (B) moves the score from 11 to 24 of 28. Three changes, in order of value:
   - Do not expose `dir` in the MCP schema. sisyphus-mcp already runs in the repo.
   - If `dir` stays, do not name the tool in the description. "Defaults to sisyphus-mcp's own working directory" causes `"dir": "sisyphus-mcp"` in 15 of 28 first calls. Use: "Leave empty. Only set it to work on a different repo."
   - Make the `chdir` error say what to do: `chdir sisyphus-mcp: no such file or directory. Leave dir empty to use the current repo.` The model retried with a new invented path in almost every run. It does recover from "bad argument" errors when the message names the fix (t12 in C).
   - Confirms qwen3:4b idea 8 and qwen3:8b idea 5, and makes it the strongest idea for this model. Weakened for qwen3.5:9b and ministral-3:8b. The description fix is new.
2. **Remove the nested example from the name descriptions, and reject a new name below an issue name.** Evidence: 4 of 4 B and C t13 runs create `auth-epic/audit-session-tokens`, 3 of them without `parent` (3 failed runs). A t06 creates `auth/login-rate-limit` (2 runs). A t08 guesses `theme/update`, B t08 rep1 guesses `docs-site/theme-update`. The tested binary has "for example fix-login-bug or web/auth/fix-login-bug" (spike workspace `cmd/sisyphus-mcp/tools.go` line 93, and 4 more `web/auth/fix-login-bug` examples on lines 139, 172, 248, 268). The main checkout at `53f4d7f` no longer has it. Merge that change into the tested binary and run t13 again. Also make `sisyphus_new` return an error such as `'auth-epic' is an issue, not a directory. To make a sub-issue, use name audit-session-tokens with parent auth-epic.`
   - Confirms ministral-3:8b ideas 1 and 2. qwen3:14b is the second model where this example costs t13.
3. **Add `bookmark` to list and search rows, and add owner, state, and bookmark to the text output of search.** Evidence: t03 fails in all 4 B and C runs. B reads the text search row (no owner, no bookmark) and says "owner is not set, no bookmark". C reads the JSON row (owner, no bookmark) and stops with an empty answer.
   - Confirms qwen3.5:9b idea 1 and ministral-3:8b idea 4. For qwen3:14b, the text form also hides `owner`, so it should get the same fields as JSON.
4. **Match search words in any order, and match the issue name.** Evidence: 8 empty searches for an issue that exists (B t08 rep0: 3, B t08 rep1: 1, C t08 rep0: 2, C t08 rep1: 2). Queries: "docs site theme update" (7 calls) and "docs-site-theme-update" (1). These are the direct cause of 3 failed t08 runs (B rep0, C rep0, C rep1).
   - Confirms qwen3:4b idea 1, qwen3:8b idea 3, qwen3.5:9b idea 2, and ministral-3:8b idea 5.
5. **Say "no results" in words, and name the next step.** Evidence: after empty searches, B t08 rep0 creates a new issue, and C t08 asks the user to "create it first using `sisyphus_new`". A result such as `No issues match "docs site theme update" in states open. Try fewer words, other states, or sisyphus_list to see all issues.` would push the model to browse. qwen3:14b reads `[]` correctly (t05 6 of 6), so the gain is the next-step hint, not the "no results" text.
   - Confirms qwen3:8b idea 2 with a new reason. Weakened as an anti-hallucination fix: qwen3:14b never invented issues.
6. **Suggest close names when a name is not found.** Evidence: 11 `No issue named` errors (2, 7, 2). B and C recover in t09 by a search. In t08 the guess and the search both fail. A message with `Similar issues: update-docs-site (Update the docs site theme)` would fix t08 at once.
   - Confirms qwen3:8b idea 1 and ministral-3:8b idea 3. Medium evidence: qwen3:14b retries more than qwen3:8b, but its retries use the same bad phrase.
7. **Make `state` optional in `sisyphus_update`.** Evidence: 4 missing-state errors (A t12 both runs, C t12 both runs). C recovers in both runs, but each costs 2 extra calls and about 400-800 tokens. C t09 rep0 invents `current_state` from the prompt rule.
   - Confirms the earlier notes. Weak for qwen3:14b in B (0 errors).
8. **Make the result of a write say what changed.** Evidence: C t13 says "created under the `auth-epic` directory" with no parent link. C t10 says "has been created" for a parent change. A write result such as `Created audit-session-tokens (parent: none, depends on: fix-login-timeout)` shows the missing parent in the same turn.
   - Confirms ministral-3:8b idea 8. Weak.
9. **Do not change:** `dependsOn` on `sisyphus_new` (correct in all 6 B and C t13 runs), `sisyphus_parent` and `sisyphus_depends_on` (t10, t11 8 of 8 in B and C), and the `unexpected additional properties` error (the model fixes the call at once).

## Verdict

- qwen3:14b obeys "think off". With the detailed prompt and no `dir` (B), it is good and fast: 24/28 (85.7%) at 1.4 s per task and 84 output tokens. It never invents issues after an empty result. It never leaks reasoning.
- With the basic setup (A), it is the worst config in the whole test: 11/28 (39%). The cause is one argument. The model sends `dir` in 70 of 76 calls, copies "sisyphus-mcp" from the `dir` description, invents paths, and asks the user for the path. Hiding `dir` fixes it.
- Thinking (C) makes it worse: 21/28 (75%), multi-step 1/4, at 9.0 s per task. The cause is not runaway thinking (max 1,385 tokens). The model stops early: 3 empty final answers and 1 skipped write. It also asks the user in t08 instead of a `sisyphus_list` call. The t13 parent goes into the name in both B and C.
- Best config: **B**. 24/28 at 1.4 s. Its 4 failures: t03 (both runs, missing fields in search rows), t08 rep0 (stray new issue), t13 rep1 (parent as directory). sisyphus-mcp ideas 2, 3, and 4 address all of them.
- Against the other local models:
  - qwen3.5:9b C is better: 28/28 against 24/28, and 5.83 GB VRAM against 10.47 GB. The only cost is time: 3.3 s against 1.4 s per task.
  - qwen3:14b B ties qwen3:4b C and qwen3:8b C (24/28), and beats qwen3:8b B (23/28) by one run, but it needs 4.2 GB more VRAM than qwen3:8b and makes 35% fewer tokens per second. qwen3:8b B is faster (0.9 s against 1.4 s).
  - qwen3:14b is safer than qwen3:8b B on read tasks (0 invented issues against 3), but it has a new risk: a stray `sisyphus_new` when it cannot find an issue.
  - The bigger model does not help with the hard tasks. t03, t08, and t13 fail as often as with the smaller qwen3 models or more often.
- Do not use qwen3:14b as the default local task manager. Use qwen3.5:9b C. If qwen3:14b is used, use B, never think on, and never expose `dir`. With 2 repeats per task, a difference of 1-3 runs between configs is near noise. The A-to-B gain (13 runs) and the `dir` cause are not noise.
