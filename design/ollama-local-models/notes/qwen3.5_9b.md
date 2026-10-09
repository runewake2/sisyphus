# qwen3.5:9b as a sisyphus task manager

## Facts

- Model: qwen3.5:9b, 9.0B parameters, Q4_K_M, family `qwen35`. Ollama 0.40.2 on a Windows workstation GPU with 16 GB VRAM.
- Capabilities from Ollama: completion, vision, tools, thinking.
- Loaded size: 5,828,646,010 bytes (5.83 GB). All of it is in VRAM. This is 0.47 GB less than qwen3:8b (6.30 GB), although the model has more parameters.
- Settings: num_ctx 8192, temperature 0.2, max_turns 8, 2 repeats per task, seed 7 and 8.
- Suite: 14 tasks. t01-t05 read, t06-t12 write, t13-t14 multi-step. Each run starts from a fresh repo with 6 seed issues.
- All 13 sisyphus-mcp tools are given to the model ("tools all").
- Configs:
  - A: think off, prompt "basic", schema raw (the optional `dir` argument is visible).
  - B: think off, prompt "detailed", schema nodir (`dir` is hidden and removed).
  - C: think on, prompt "detailed", schema nodir.
- The checker fix (state from the state directory, not the parent directory) was in place for all three configs. No re-scoring is necessary. The fix matters here: all 6 t06 runs created a nested name (see "Failure modes"), and the old checker would fail them.
- The tested `sisyphus-mcp` binary comes from the `sisyphus-ollama-model-spike` workspace. Its `name` descriptions contain the nested example `web/auth/fix-login-bug`. The `sisyphus` checkout at HEAD does not have this example.
- **qwen3.5:9b obeys "think off".** 0 of 56 think-off final answers contain `<think>`, `</think>`, or reasoning text ("Let me", "The user wants", "Wait,"). The answers are short: 118 characters on average in A (max 196), 103 in B (max 199). C answers are also clean: 131 characters on average (max 481, a t05 answer that lists the open issues).
- The model never answered without a tool call (no_tool_runs = 0 in all configs). It never called an unknown tool. No final answer is empty. No run hit max_turns (max 7 turns, B t08 rep0).
- The model never sent `dir`, also in A where `dir` is visible. It never called `sisyphus_slug`.
- **The model never guessed an issue name.** There is 0 `No issue named` error in 84 runs. qwen3:8b had 22 such errors, qwen3:4b had 23. Every name in a `sisyphus_update`, `sisyphus_show`, `sisyphus_parent`, or `sisyphus_depends_on` call came from the user prompt or from a tool result.
- No warm-up outlier. The first run (A t01 rep0) took 1.57 s. Total load time is 0.7 s per config.
- The harness does not save the thinking text. For C, the thinking size is only visible as output tokens. An estimate of the thinking part is "C tokens minus B tokens" for the same task.

## Results

| | A | B | C |
|---|---|---|---|
| Pass rate | **23/28 (82.1%)** | **24/28 (85.7%)** | **28/28 (100%)** |
| Read | 8/10 | 8/10 | 10/10 |
| Write | 13/14 | 12/14 | 14/14 |
| Multi | 2/4 | 4/4 | 4/4 |
| Mean check score | 0.907 | 0.923 | 1.000 |
| Generation speed | 115.4 tok/s | 113.9 tok/s | 115.1 tok/s |
| Prompt speed (cached prefix) | 28,635 tok/s | 27,638 tok/s | 22,867 tok/s |
| Mean wall time per task | 1.61 s | 1.46 s | 3.28 s |
| Median wall time per task | 1.45 s | 1.15 s | 3.32 s |
| Max wall time per task | 3.60 s | 3.24 s | 5.72 s |
| Mean output tokens per task | 129 | 112 | 309 |
| Median output tokens per task | 101 | 87 | 305 |
| Mean wall time, passed / failed runs | 1.55 s / 1.88 s | 1.37 s / 2.04 s | 3.28 s / none |
| Tool calls (per run) | 46 (1.64) | 47 (1.68) | 49 (1.75) |
| Mean turns per run | 2.64 | 2.68 | 2.75 |
| Tool-call errors | 4 | 2 | 0 |
| Runs with an error that then passed | 2 of 3 | 0 of 2 | none |
| Runs with no tool call | 0 | 0 | 0 |
| Runs that hit max_turns | 0 | 0 | 0 |
| Final answers with leaked reasoning | 0 | 0 | 0 |
| Loaded VRAM | 5.83 GB | 5.83 GB | 5.83 GB |

Tool-call errors by type:

| Error | A | B | C |
|---|---|---|---|
| `No issue named '...'` (guessed name) | 0 | 0 | 0 |
| `chdir ...: no such file or directory` (junk `dir`) | 0 | hidden | hidden |
| `missing properties: ["state"]` (update without state) | 3 | 0 | 0 |
| `Issue 'fix-login-timeout' already exists` (`sisyphus_new` to start work) | 1 | 2 | 0 |
| `unexpected additional properties` | 0 | 0 | 0 |

Per kind, mean wall time and output tokens:

| Kind | A | B | C |
|---|---|---|---|
| Read | 1.05 s, 69 tok | 0.97 s, 62 tok | 2.79 s, 259 tok |
| Write | 1.91 s, 159 tok | 1.79 s, 142 tok | 3.42 s, 320 tok |
| Multi | 1.91 s, 170 tok | 1.53 s, 128 tok | 3.98 s, 391 tok |

### Output tokens in C (thinking included)

| | qwen3.5:9b C | qwen3:8b C | qwen3:4b C |
|---|---|---|---|
| Min | 131 | 360 | 565 |
| Median | 305 | 677 | 2,066 |
| p90 | 437 | 1,321 | 5,381 |
| Max | 564 (t07 rep0) | 16,490 (t12 rep1) | 8,944 (t12 rep1) |
| Mean | 309 | 1,328 | 2,621 |
| Max output tokens in one turn | 280 | 8,245 | 2,928 |
| Mean / max wall time | 3.3 s / 5.7 s | 11.3 s / 140.7 s | 14.9 s / 51.7 s |

- There is no runaway thinking. The largest run uses 564 tokens. The largest single turn uses 280 tokens. The distribution is narrow: p90 is only 1.4 times the median. For qwen3:8b, the max is 24 times the median.
- Estimated thinking per run (C minus B, same task): 197 tokens on average. The range per task is 90 (t02) to 428 (t05). t05 is high because C does 3 to 5 searches before it says "no".
- Prompt per turn is 3.9k to 4.1k tokens on average, max 4.5k (C t08 rep1, after a 6 KB search result). The context is never a problem.

## Tasks

Passes per task, out of 2:

| Task | Kind | A | B | C | Note |
|---|---|---|---|---|---|
| t01 high/critical open issues | read | 2 | 2 | 2 | |
| t02 timeout issue | read | 2 | 2 | 2 | |
| t03 docs owner and bookmark | read | **0** | **0** | 2 | think off answers the issue name as the bookmark |
| t04 sub-issues of auth-epic | read | 2 | 2 | 2 | |
| t05 payment issue (none) | read | 2 | 2 | 2 | reads `[]` correctly in all 6 runs |
| t06 create rate-limit issue | write | 2 | 2 | 2 | all 6 runs create a nested name |
| t07 start fix-login-timeout | write | 1 | **0** | 2 | think off keeps state open |
| t08 close docs theme issue | write | 2 | 2 | 2 | first model with 6/6 |
| t09 abandon password reset | write | 2 | 2 | 2 | |
| t10 set parent | write | 2 | 2 | 2 | |
| t11 add dependency | write | 2 | 2 | 2 | |
| t12 lower priority | write | 2 | 2 | 2 | A recovers from the missing-state error |
| t13 create with parent and dependency | multi | 2 | 2 | 2 | one call, never nested |
| t14 find critical, start it, give title | multi | **0** | 2 | 2 | A keeps state open |

Side by side with qwen3:4b and qwen3:8b (qwen3:4b uses the corrected scores):

| Task | 4b A | 4b B | 4b C | 8b A | 8b B | 8b C | 3.5:9b A | 3.5:9b B | 3.5:9b C |
|---|---|---|---|---|---|---|---|---|---|
| t01 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t02 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t03 | 2 | 1 | 2 | 0 | 0 | 2 | 0 | 0 | 2 |
| t04 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t05 | 2 | 2 | 2 | 0 | 0 | 2 | 2 | 2 | 2 |
| t06 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t07 | 2 | 2 | 2 | 2 | 2 | 2 | 1 | 0 | 2 |
| t08 | 0 | 0 | 0 | 0 | 1 | 0 | 2 | 2 | 2 |
| t09 | 1 | 1 | 2 | 0 | 2 | 1 | 2 | 2 | 2 |
| t10 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t11 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t12 | 2 | 2 | 1 | 0 | 2 | 1 | 2 | 2 | 2 |
| t13 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 | 2 |
| t14 | 0 | 2 | 1 | 2 | 2 | 2 | 0 | 2 | 2 |
| **Total** | 23 | 24 | 24 | 18 | 23 | 24 | 23 | 24 | **28** |
| Mean wall time | 14.4 s | 16.7 s | 14.9 s | 1.2 s | 0.9 s | 11.3 s | 1.6 s | 1.5 s | 3.3 s |
| Mean output tokens | 2,542 | 3,000 | 2,621 | 90 | 90 | 1,328 | 129 | 112 | 309 |
| Generation speed | 187 tok/s | 187 tok/s | 183 tok/s | 117 tok/s | 119 tok/s | 123 tok/s | 115 tok/s | 114 tok/s | 115 tok/s |
| VRAM | 3.87 GB | 3.87 GB | 3.87 GB | 6.30 GB | 6.30 GB | 6.30 GB | 5.83 GB | 5.83 GB | 5.83 GB |

- t08 was the hardest task for the qwen3 models (1 pass in 12 runs). qwen3.5:9b passes it in all 6 runs. See "How qwen3.5:9b finds names".
- t03 fails with think off for a new reason. qwen3:8b guessed a name. qwen3.5:9b finds the correct issue, but the tool output has no bookmark. See "Why t03 fails with think off".
- t07 is a new failure. qwen3:4b and qwen3:8b passed it in all 12 runs.

### Why t03 fails with think off

The task: "Who owns the docs site issue, and what bookmark is it on?"

- All 4 think-off runs (A and B) make one call: `sisyphus_search {"query": "docs site"}`. It finds the correct issue at once.
- The search row has `name`, `title`, `state`, `priority`, `owner`, `tags`, and `content` (the body). It has **no `bookmark` field**. The bookmark `ai/docs-theme` is not in the body either.
- The model then answers with the issue name as the bookmark. All 4 runs give almost the same answer: "The docs site issue is owned by **alice** and is on bookmark **update-docs-site**." The owner is correct. The bookmark is invented.
- So the model does not omit a value that it has. The tool output does not contain the value, and the model fills the gap instead of making one more call.
- This is a sisyphus-mcp output issue. `listRow` (`cmd/sisyphus/list.go`) and `searchRow` (`cmd/sisyphus/search.go`) include `owner` but not `bookmark`. A row that shows `owner` but not `bookmark` suggests that the issue has no bookmark. Only `sisyphus_show` returns the bookmark.
- C does the same search, and then calls `sisyphus_show {"name": "update-docs-site"}`. The show output has `"bookmark": "ai/docs-theme"`. Both C runs pass with 2 calls, 263-269 tokens, and 2.9 s.
- The qwen3:4b notes have one failed run of the same kind: B t03 rep0 said the bookmark is "not listed". That is probably the same cause.

### How qwen3.5:9b finds names (t08 and others)

The task t08: "The docs site theme update is done. Close it." The issue is `update-docs-site`, title "Update the docs site theme". The search is a phrase match on title and body, so "docs site theme update" returns `[]`.

- The model never tries a write with a guessed name. Its first call is always a search with the user's words.
- After an empty result, it changes the query or it browses. It does not ask the user. Every run:
  - A rep0: search "docs site theme update" `[]`, search "docs theme" `[]`, `sisyphus_list {"state": "open", "text": true}`, `sisyphus_list {"state": "in-progress", "text": true}` (finds it), `sisyphus_update {"name": "update-docs-site", "state": "closed"}`. 5 calls.
  - A rep1: the same steps, JSON form. 5 calls.
  - B rep0: two empty searches, `sisyphus_list {"tags": "feature"}`, `sisyphus_list {"tags": "bug"}`, `sisyphus_list {}` (finds it), update. 6 calls.
  - B rep1: search "docs site theme update" `[]`, search "theme update docs" `[]`, `sisyphus_list {"tags": "feature"}`, search "theme" (finds it), update. 5 calls.
  - C rep0: search "docs site theme update" `[]`, search "theme" (finds it), update with resolution completed. 3 calls, 329 tokens, 3.6 s.
  - C rep1: search "docs site theme update" `[]`, search "docs theme" `[]`, search "docs" (returns 5 issues, see idea 3), update. 4 calls, 347 tokens, 4.2 s.
- C finds the issue with fewer calls (3.5 against 5.25 with think off). It drops to one strong word ("theme") sooner. The think-off configs browse with `sisyphus_list`. Both ways work.
- t09 shows the same pattern. B searches "password reset feature" (`[]`), then "password reset" (found). A and C search "password reset" at once.
- t07 and t12 in C start with `sisyphus_show` on the name from the prompt. This confirms that the issue exists and gives the current state before the write. So C has 0 tool errors.

Calls and output tokens per task in C, against qwen3:8b C (2 runs each):

| Task | qwen3.5:9b calls | qwen3.5:9b tokens | qwen3:8b calls | qwen3:8b tokens |
|---|---|---|---|---|
| t01 | 1, 1 | 193, 193 | 1, 1 | 819, 665 |
| t02 | 1, 1 | 131, 133 | 1, 1 | 718, 410 |
| t03 | 2, 2 | 263, 269 | 3, 3 | 904, 993 |
| t04 | 1, 1 | 231, 238 | 1, 1 | 595, 449 |
| t05 | 3, 5 | 387, 554 | 1, 1 | 377, 360 |
| t06 | 1, 1 | 366, 379 | 1, 1 | 879, 568 |
| t07 | 2, 2 | 564, 354 | 1, 1 | 848, 803 |
| t08 | 3, 4 | 329, 347 | 1, 1 (fail) | 647, 651 |
| t09 | 2, 2 | 286, 287 | 3, 1 (1 fail) | 1,497, 1,246 |
| t10 | 1, 1 | 227, 232 | 1, 1 | 591, 597 |
| t11 | 1, 1 | 238, 203 | 1, 1 | 533, 574 |
| t12 | 2, 2 | 347, 325 | 3, 1 (1 fail) | 1,616, 16,490 |
| t13 | 1, 1 | 560, 338 | 1, 1 | 688, 595 |
| t14 | 2, 2 | 342, 323 | 2, 2 | 957, 1,103 |
| **Total** | 49 calls | 8,639 tok | 38 calls | 37,173 tok |

- qwen3.5:9b makes more calls (49 against 38) but thinks much less per call. It spends tokens on tool calls, not on thinking. qwen3:8b in t08 thought about 650 tokens, made one wrong call, and asked the user. qwen3.5:9b used fewer tokens and made 3-4 calls.

## Failure modes

Failed runs: A 5, B 4, C 0. One run can show more than one mode. The table counts failed runs that show the mode. The numbering follows the qwen3:4b and qwen3:8b notes. Modes 17-19 are new for qwen3.5:9b.

| Mode | A | B | C |
|---|---|---|---|
| 1. Guessed issue name (not taken from a tool result) | 0 | 0 | 0 |
| 2. Search that returns nothing (phrase match) | 0 | 0 | 0 |
| 3. `sisyphus_slug` used to "find" an existing issue | 0 | 0 | 0 |
| 4. Wrong final answer to a read task | 2 | 2 | 0 |
| 5. Skips the write step, then answers only the question part | 0 | 0 | 0 |
| 6. Unknown argument | 0 | 0 | 0 |
| 7. Junk `dir` argument | 0 | n/a | n/a |
| 8. Loop until max_turns | 0 | 0 | 0 |
| 9. Wrong tool | 1 | 2 | 0 |
| 10. Repeats the same call with no change | 0 | 0 | 0 |
| 11. Stray mutation | 0 | 0 | 0 |
| 12. Answered without tools | 0 | 0 | 0 |
| 13. Gives up and asks the user | 0 | 0 | 0 |
| 14. Invents issue data after an empty tool result | 0 | 0 | 0 |
| 15. `sisyphus_update` without `state` | 1 | 0 | 0 |
| 16. Runaway thinking | 0 | 0 | 0 |
| 17. **Invents a field that the tool output does not have** (bookmark = issue name) | 2 | 2 | 0 |
| 18. **"Start" done with state `open`, not `in-progress`** | 3 | 2 | 0 |
| 19. **`sisyphus_new` on an existing issue to start work** (mode 9 above) | 1 | 2 | 0 |

Every failed run per config:

- A: t03 rep0 and rep1 (4, 17), t07 rep1 (9, 15, 18, 19), t14 rep0 and rep1 (18).
- B: t03 rep0 and rep1 (4, 17), t07 rep0 and rep1 (9, 18, 19).
- C: none.

Examples:

17. Invents a field that the tool output does not have. All 4 think-off t03 runs:
   - `sisyphus_search {"query": "docs site"}` returns one row with `"owner": "alice"` and no bookmark.
   - Final (A rep0, B rep0, B rep1): "The docs site issue is owned by **alice** and is on bookmark **update-docs-site**."
   - This differs from qwen3:8b mode 14. qwen3:8b invented whole issues after an empty result. qwen3.5:9b reads empty results correctly (t05 passes in 6 of 6 think-off runs: "No, there is no issue about payment processing"). It only fills one missing field.
18. "Start" done with state `open`. The model sets owner and bookmark, but keeps the state:
   - A t14 rep0: `sisyphus_update {"name": "migrate-database", "state": "open", "owner": "carol", "bookmark": "ai/db"}`. Final: "I've started it with owner carol and bookmark ai/db." The answer says "started", but the issue is still open. Both A t14 runs do this.
   - B t07 rep1: final "It remains in the \"open\" state." The model knows that it did not change the state.
   - The detailed prompt has the line "To start work: sisyphus_update with state in-progress". It fixes t14 (A 0/2 to B 2/2), but not t07 (B 0/2). In t07, the model goes to `sisyphus_new` first (mode 19), and after that error it keeps `"state": "open"` from its first call.
19. `sisyphus_new` on an existing issue. "Start work on fix-login-timeout" is read as "create":
   - B t07 rep0: `sisyphus_new {"name": "fix-login-timeout", "title": "Fix login timeout issue", "state": "open", "priority": "medium", "effort": "medium", "tags": "", "owner": "bob", "bookmark": "ai/login-timeout"}` returns `Error: Issue 'fix-login-timeout' already exists: issues/open/fix-login-timeout.md.` Then `sisyphus_show`, then `sisyphus_update {"name": "fix-login-timeout", "state": "open", "owner": "bob", "bookmark": "ai/login-timeout"}`.
   - The `already exists` error does not say which tool changes an existing issue.
   - `sisyphus_new` accepts `owner` and `bookmark`, so the call looks valid to the model.
15. Update without state:
   - A t12 rep0 and rep1, A t07 rep1: `missing properties: ["state"]`. The model recovers in all 3 cases. In t12 rep1 it calls `sisyphus_show` first to read the current state. In t07 rep1 it recovers with the wrong state (mode 18).
   - With the detailed prompt (B, C), this error does not occur.

Other observations (no failed run):

- **Nested names on create.** All 6 t06 runs create a nested name. The user did not ask for directories:
  - A: `api/login/add-rate-limiting`, `auth/fix-login-rate-limiting`.
  - B: `auth/fix-login-rate-limiting` (both runs).
  - C: `auth/rate-limiting-login-api`, `auth/rate-limit-login-api`.
  - The checker accepts these because it reads the state directory. qwen3:8b never created a nested name. qwen3:4b did it in t13 only. qwen3.5:9b never nests in t13, where the user gives the name.
- Empty searches for issues that exist: A 4, B 6, C 3 (13 in total). Queries: "docs site theme update" (6), "docs theme" (4), "theme update docs" (1), "password reset feature" (2). All of them recover with a shorter query or a list.
- C t06 rep0 gives no `title`. The title comes from the name. The check passes.

## Effect of settings

### A to B: detailed prompt and hidden `dir`

- Pass rate: 23 to 24 of 28.
- Fixed:
  - t14 (0/2 to 2/2). The prompt line "To start work: sisyphus_update with state in-progress" works here.
  - The missing-state error (3 to 0). The prompt line "pass the current state" works. t12 passed in A also, but with a retry.
- Broke: t07 (1/2 to 0/2). B starts both runs with `sisyphus_new` (mode 19), then keeps state `open`. A rep0 had used `sisyphus_update` with `in-progress` at once. With 2 repeats, this can be noise.
- No change: t03 (0/2). The prompt cannot add a bookmark to the search output.
- Hiding `dir` had no effect. A never sent `dir`.
- Tool calls: 46 to 47. Tool-call errors: 4 to 2.
- Cost: none. Mean wall time 1.61 to 1.46 s. Mean output tokens 129 to 112.

### B to C: thinking on

- Pass rate: 24 to 28 of 28.
- Fixed:
  - t03 (0/2 to 2/2). C calls `sisyphus_show` after the search and reads the bookmark.
  - t07 (0/2 to 2/2). C calls `sisyphus_show` first, then `sisyphus_update` with `in-progress`. No `sisyphus_new`.
- Broke: nothing.
- C checks before it writes. It calls `sisyphus_show` 6 times (B: 2 times, both after an error). It uses `sisyphus_search` 17 times (B: 15).
- Tool calls: 47 to 49. Tool-call errors: 2 to 0.
- Cost: small. Mean wall time 1.46 to 3.28 s (2.2 times). Median 1.15 to 3.32 s. Mean output tokens 112 to 309 (2.8 times). For qwen3:8b, the same step cost 12 times the time and 15 times the tokens.
- No runaway. Max 564 tokens and 5.7 s per run.

## Ideas for sisyphus-mcp

The ideas are in order of evidence for qwen3.5:9b. For each idea, the last line says if it confirms, weakens, or is new compared with the qwen3:4b and qwen3:8b notes.

1. **Add `bookmark` to the rows of `sisyphus_list` and `sisyphus_search`.** Evidence: all 4 think-off t03 runs find the correct issue with one search, see `owner` but no `bookmark`, and answer the issue name as the bookmark. C needs one more call (`sisyphus_show`) for the same answer. Add `bookmark` to `listRow` and `searchRow` (also the `bookmark` column in the text table), with `omitempty` as for `owner`. Consider `parent` and `resolution` also. A row that has some fields but not others makes a model think that the missing field is empty.
   - New. It probably also explains qwen3:4b B t03 rep0 ("bookmark not listed").
2. **Match search words in any order, and match the issue name.** Evidence: 13 empty searches for issues that exist. "docs theme" and "theme update docs" return `[]`, but the title is "Update the docs site theme". "password reset feature" returns `[]`. The search is one lowercase substring match (`strings.Contains` in `cmd/sisyphus/search.go`). Split the query into words (on spaces, `-`, and `/`), and match each word against name, title, and body. qwen3.5:9b recovers each time, but it costs 1 to 4 extra calls (t08 needs 3 to 6 calls instead of 2).
   - Confirms qwen3:4b idea 1 and qwen3:8b idea 3.
3. **Do not match the template placeholder text in search, and return less body text.** Evidence: C t08 rep1 searched "docs". It returned all 5 non-closed issues, because each body has the template placeholder "Link to design docs ...". The result was 6,086 characters, over the harness cut of 6,000. Every search row returns the full body as `content`, mostly placeholder text. Ignore lines in `<...>` placeholders when matching. Return a short snippet around the match, or only the Summary section, unless the caller asks for the full body.
   - New.
4. **Remove the nested example from the `sisyphus_new` name description.** Evidence: all 6 t06 runs create a nested name (`auth/fix-login-rate-limiting` and others). The description says "optionally below kebab-case directories, for example fix-login-bug or web/auth/fix-login-bug". The model copies the `web/auth/` pattern into `auth/` and `api/login/`. Say "for example fix-login-bug" only, and describe directories in the docs, not in the schema.
   - Confirms qwen3:4b idea 3 and qwen3:8b idea 6, and makes it stronger for create. qwen3:8b only nested on lookup; qwen3.5:9b nests on create in 6 of 6 runs.
5. **Make the "already exists" error name the next tool, and tie "start" to `in-progress`.** Evidence: 3 runs call `sisyphus_new` on an existing issue for "Start work on fix-login-timeout", and 5 failed runs set owner and bookmark with state `open`. Two changes:
   - Error: `Issue 'fix-login-timeout' already exists (state open). To change it, use sisyphus_update. To start work, use state in-progress.`
   - `sisyphus_update` `state` description: "The new state: open, in-progress (work has started), or closed." The `bookmark` description already says "Set it when work starts", but it does not say which state.
   - New.
6. **Make `state` optional in `sisyphus_update`.** Evidence: 3 missing-state errors in A, 0 in B and C. The model recovers in all 3 cases, but in t07 rep1 it recovers with the wrong state.
   - Confirms qwen3:4b idea 5 and qwen3:8b idea 4, but weaker. qwen3.5:9b recovers by itself.
7. **Suggest close names when a name is not found.** Evidence: none for qwen3.5:9b. It never guessed a name (0 `No issue named` errors).
   - Weakens qwen3:8b idea 1 for this model. Keep it for smaller models.
8. **Say "no results" in words for an empty result.** Evidence: none. qwen3.5:9b reads `[]` correctly in all 6 t05 runs (and in 13 other empty searches).
   - Weakens qwen3:8b idea 2 for this model. Keep it for qwen3:8b.
9. **Hide or drop `dir` for agents.** Evidence: none. A never sent `dir`.
   - Weakens qwen3:4b idea 8 and qwen3:8b idea 5 for this model.
10. **Say what `sisyphus_slug` is for.** Evidence: none. No call.
    - Same as qwen3:8b.
11. **Do not change:** `parent` and `dependsOn` on `sisyphus_new`. All 6 t13 runs use them in one call with the exact name.

## Verdict

- qwen3.5:9b is the best local model in this test. It obeys "think off" (0 leaks), it never guesses issue names, and it never gives up and asks the user. When a search is empty, it changes the query or browses with `sisyphus_list`.
- Best config: **C** (think on, detailed prompt, no `dir`). 28/28 (100%) at 3.3 s mean and median, 309 output tokens per task, max 5.7 s and 564 tokens. No tool error, no runaway thinking. Thinking costs only about 200 tokens per task.
- Fast config: **B** (think off, detailed prompt, no `dir`). 24/28 (85.7%) at 1.5 s. Risks: t03 gives a confident wrong bookmark (4 of 4 think-off runs), and "start work" can leave the issue open (t07). Ideas 1 and 5 address both failures directly. After these fixes, B can probably reach 27-28/28.
- Against qwen3:4b and qwen3:8b:
  - Best pass rate: 28/28 against 24/28 for both.
  - Best-config speed: 3.3 s against 11.3 s (qwen3:8b C) and 14.9 s (qwen3:4b C). The tail is much shorter: max 5.7 s against 140.7 s and 51.7 s.
  - qwen3.5:9b C is 3.4 times faster than qwen3:8b C for 4 more passes in 28. Only qwen3:8b B (0.9 s, 23/28) is faster, and it can invent issues.
  - VRAM: 5.83 GB, less than qwen3:8b (6.30 GB). Generation speed is similar (115 against 120 tok/s). The speed gain comes from fewer tokens.
  - New weak points, not seen in qwen3: nested names on create (6 of 6 t06 runs) and "start" without `in-progress` with think off.
- Use qwen3.5:9b with config C as the default local task manager. With 2 repeats per task, 28/28 does not prove a 100% rate. Make ideas 1 to 5 in sisyphus-mcp, then run B again to see if think off is good enough.
