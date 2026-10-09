#!/usr/bin/env python3
"""Run an Ollama model against sisyphus-mcp on a fixed task suite.

Each task gets a fresh scratch repo seeded with the same issues. The model talks
to sisyphus-mcp through Ollama's /api/chat tool calling. After the chat ends, a
checker reads the issue files on disk (and the final answer) and scores the task.

Usage:
  harness.py --model qwen3:4b [--think off|on|default] [--repeats 2]
             [--prompt basic|none|detailed] [--tools all|core]
             [--temperature 0.2] [--num-ctx 8192] [--num-thread 16]
             [--host http://127.0.0.1:11434] [--tasks t01,t02] [--out results/]
"""
import argparse, json, os, re, shutil, subprocess, sys, tempfile, time, urllib.request
from pathlib import Path

import yaml

HERE = Path(__file__).resolve().parent
# BENCH_WORK: work directory for scratch repos, results, and the PAUSE switch.
# SISYPHUS_BIN: directory that holds the sisyphus and sisyphus-mcp binaries (default: found on PATH).
SCRATCH = Path(os.environ.get("BENCH_WORK", HERE.parent / "work"))
BIN = Path(os.environ.get("SISYPHUS_BIN") or Path(shutil.which("sisyphus") or "sisyphus").parent)

# ---------------------------------------------------------------- seed repo

SEED = [
    # name, state, fields, summary
    ("auth-epic", "open", {"title": "Rework authentication", "priority": "high", "tags": ["auth"]},
     "Rework the authentication flow. This epic holds the auth sub-issues."),
    ("fix-login-timeout", "open", {"title": "Fix login timeout on slow networks", "priority": "high",
                                   "tags": ["bug", "auth"], "parent": "auth-epic"},
     "Login requests fail with a timeout after 5 seconds on slow mobile networks."),
    ("add-password-reset", "open", {"title": "Add password reset by email", "priority": "medium",
                                    "tags": ["feature", "auth"], "parent": "auth-epic"},
     "Users cannot reset a forgotten password. Send a reset link by email."),
    ("update-docs-site", "in-progress", {"title": "Update the docs site theme", "priority": "low",
                                         "tags": ["docs"], "owner": "alice", "bookmark": "ai/docs-theme"},
     "Move the docs site to the new theme."),
    ("migrate-database", "open", {"title": "Migrate to Postgres 17", "priority": "critical", "tags": ["db"]},
     "Upgrade the production database from Postgres 15 to Postgres 17."),
    ("old-cache-layer", "closed", {"title": "Remove the old cache layer", "priority": "medium",
                                   "tags": ["cache"], "resolution": "completed"},
     "Delete the unused Redis cache layer."),
]


def sis(repo, *args):
    r = subprocess.run([str(BIN / "sisyphus"), *args], cwd=repo, capture_output=True, text=True)
    if r.returncode:
        raise RuntimeError(f"sisyphus {' '.join(args)}: {r.stderr}")
    return r.stdout


def make_repo():
    (SCRATCH / "runs-tmp").mkdir(parents=True, exist_ok=True)
    repo = Path(tempfile.mkdtemp(prefix="sisrepo-", dir=SCRATCH / "runs-tmp"))
    sis(repo, "init")
    for name, state, f, summary in SEED:
        args = ["new", name, "--title", f["title"], "--priority", f["priority"],
                "--tags", ",".join(f["tags"]), "--state", "open"]
        if "parent" in f:
            args += ["--parent", f["parent"]]
        sis(repo, *args)
        if state == "in-progress":
            sis(repo, "update", name, "in-progress", "--owner", f["owner"], "--bookmark", f["bookmark"])
        elif state == "closed":
            sis(repo, "update", name, "closed", "--resolution", f["resolution"])
        p = find_issue(repo, name)
        txt = p.read_text().replace("<One or two sentences. Tell what must change and why.>", summary)
        p.write_text(txt)
    return repo


def find_issue(repo, name):
    for st in ("open", "in-progress", "closed"):
        p = Path(repo) / "issues" / st / f"{name}.md"
        if p.exists():
            return p
    return None


def snapshot(repo):
    out = {}
    for p in sorted((Path(repo) / "issues").rglob("*.md")):
        if p.name == "TEMPLATE.md":
            continue
        txt = p.read_text()
        fm = yaml.safe_load(txt.split("---", 2)[1]) or {}
        # State dir, not the parent: names may nest below it, for example issues/open/web/auth/fix-login-bug.md
        fm["_dir"] = p.relative_to(Path(repo) / "issues").parts[0]
        out[p.stem] = fm
    return out


# ---------------------------------------------------------------- tasks

def has(ans, *words):
    a = ans.lower()
    return all(w.lower() in a for w in words)


def link(v):
    """Normalize '[[x]]' / 'x' / ['[[x]]'] to a set of names."""
    if v is None:
        return set()
    if isinstance(v, str):
        v = [v]
    return {re.sub(r"^\[\[|\]\]$", "", str(x)).strip() for x in v}


def changed(before, after):
    names = set(before) | set(after)
    return {n for n in names if before.get(n) != after.get(n)}


def only_changed(before, after, allowed):
    extra = changed(before, after) - set(allowed)
    return (not extra, f"unexpected changes: {sorted(extra)}" if extra else "no stray changes")


TASKS = []


def task(tid, kind, prompt):
    def deco(fn):
        TASKS.append({"id": tid, "kind": kind, "prompt": prompt, "check": fn})
        return fn
    return deco


@task("t01", "read", "Which open issues have high or critical priority? List their names.")
def _(b, a, ans):
    return [
        (has(ans, "fix-login-timeout"), "mentions fix-login-timeout"),
        (has(ans, "auth-epic"), "mentions auth-epic"),
        (has(ans, "migrate-database"), "mentions migrate-database"),
        (not has(ans, "add-password-reset"), "does not list add-password-reset (medium)"),
        (not has(ans, "update-docs-site"), "does not list update-docs-site (in-progress, low)"),
        only_changed(b, a, []),
    ]


@task("t02", "read", "Which issue is about a timeout? Give me its name.")
def _(b, a, ans):
    return [(has(ans, "fix-login-timeout"), "names fix-login-timeout"), only_changed(b, a, [])]


@task("t03", "read", "Who owns the docs site issue, and what bookmark is it on?")
def _(b, a, ans):
    return [(has(ans, "alice"), "owner alice"), (has(ans, "ai/docs-theme"), "bookmark ai/docs-theme"),
            only_changed(b, a, [])]


@task("t04", "read", "What are the sub-issues of auth-epic?")
def _(b, a, ans):
    return [(has(ans, "fix-login-timeout"), "fix-login-timeout"), (has(ans, "add-password-reset"), "add-password-reset"),
            (not has(ans, "migrate-database"), "does not list migrate-database"), only_changed(b, a, [])]


@task("t05", "read", "Is there an issue about payment processing?")
def _(b, a, ans):
    neg = re.search(r"\b(no|not|none|n't|nothing)\b", ans.lower()) is not None
    return [(neg, "answers no"), only_changed(b, a, [])]


@task("t06", "write", "Create a new issue to add rate limiting to the login API. "
                      "Make it high priority and small effort, and tag it auth and feature.")
def _(b, a, ans):
    new = [n for n in a if n not in b]
    fm = a[new[0]] if len(new) == 1 else {}
    return [
        (len(new) == 1, f"exactly one new issue (got {new})"),
        ("rate" in (fm.get("title") or "").lower() or "rate" in (new[0] if new else ""), "title/name mentions rate limiting"),
        (fm.get("priority") == "high", "priority high"),
        (fm.get("effort") == "small", "effort small"),
        ({"auth", "feature"} <= set(fm.get("tags") or []), "tags auth+feature"),
        (fm.get("_dir") == "open", "state open"),
        only_changed(b, a, new),
    ]


@task("t07", "write", "Start work on fix-login-timeout. The owner is bob and the bookmark is ai/login-timeout.")
def _(b, a, ans):
    fm = a.get("fix-login-timeout", {})
    return [(fm.get("_dir") == "in-progress" and fm.get("state") == "in-progress", "in-progress"),
            (fm.get("owner") == "bob", "owner bob"), (fm.get("bookmark") == "ai/login-timeout", "bookmark set"),
            only_changed(b, a, ["fix-login-timeout"])]


@task("t08", "write", "The docs site theme update is done. Close it.")
def _(b, a, ans):
    fm = a.get("update-docs-site", {})
    return [(fm.get("_dir") == "closed", "closed"), (fm.get("resolution") == "completed", "resolution completed"),
            only_changed(b, a, ["update-docs-site"])]


@task("t09", "write", "We decided not to build the password reset feature. Abandon that issue.")
def _(b, a, ans):
    fm = a.get("add-password-reset", {})
    return [(fm.get("_dir") == "closed", "closed"), (fm.get("resolution") == "abandoned", "resolution abandoned"),
            only_changed(b, a, ["add-password-reset"])]


@task("t10", "write", "Make migrate-database a sub-issue of auth-epic.")
def _(b, a, ans):
    fm = a.get("migrate-database", {})
    return [(link(fm.get("parent")) == {"auth-epic"}, "parent auth-epic"),
            (fm.get("_dir") == "open", "still open"), only_changed(b, a, ["migrate-database"])]


@task("t11", "write", "add-password-reset cannot start until migrate-database is done. Record that dependency.")
def _(b, a, ans):
    fm = a.get("add-password-reset", {})
    return [("migrate-database" in link(fm.get("depends-on")), "depends-on migrate-database"),
            ("add-password-reset" not in link(a.get("migrate-database", {}).get("depends-on")), "direction correct"),
            only_changed(b, a, ["add-password-reset"])]


@task("t12", "write", "Lower the priority of migrate-database to medium.")
def _(b, a, ans):
    fm = a.get("migrate-database", {})
    return [(fm.get("priority") == "medium", "priority medium"), (fm.get("_dir") == "open", "still open"),
            only_changed(b, a, ["migrate-database"])]


@task("t13", "multi", "Create an issue named audit-session-tokens with the title 'Audit session tokens' and "
                      "medium priority. Put it under auth-epic, and make it depend on fix-login-timeout.")
def _(b, a, ans):
    fm = a.get("audit-session-tokens", {})
    return [(bool(fm), "audit-session-tokens exists"), ("audit session tokens" in (fm.get("title") or "").lower(), "title"),
            (fm.get("priority") == "medium", "priority medium"), (link(fm.get("parent")) == {"auth-epic"}, "parent auth-epic"),
            ("fix-login-timeout" in link(fm.get("depends-on")), "depends-on fix-login-timeout"),
            only_changed(b, a, ["audit-session-tokens"])]


@task("t14", "multi", "Find the critical issue, start it with owner carol and bookmark ai/db, "
                      "and tell me its title.")
def _(b, a, ans):
    fm = a.get("migrate-database", {})
    return [(fm.get("_dir") == "in-progress", "migrate-database in-progress"), (fm.get("owner") == "carol", "owner carol"),
            (fm.get("bookmark") == "ai/db", "bookmark ai/db"), (has(ans, "postgres"), "answer gives title"),
            only_changed(b, a, ["migrate-database"])]


# ---------------------------------------------------------------- MCP client

class MCP:
    def __init__(self, cwd):
        env = dict(os.environ, PATH=f"{BIN}:{os.environ['PATH']}")
        self.p = subprocess.Popen([str(BIN / "sisyphus-mcp")], cwd=cwd, env=env, stdin=subprocess.PIPE,
                                  stdout=subprocess.PIPE, stderr=subprocess.DEVNULL, text=True, bufsize=1)
        self.i = 0
        self.call("initialize", {"protocolVersion": "2025-06-18", "capabilities": {},
                                 "clientInfo": {"name": "ollama-harness", "version": "0"}})
        self.notify("notifications/initialized")

    def notify(self, method, params=None):
        self.p.stdin.write(json.dumps({"jsonrpc": "2.0", "method": method, "params": params or {}}) + "\n")

    def call(self, method, params):
        self.i += 1
        self.p.stdin.write(json.dumps({"jsonrpc": "2.0", "id": self.i, "method": method, "params": params}) + "\n")
        while True:
            line = self.p.stdout.readline()
            if not line:
                raise RuntimeError("sisyphus-mcp closed")
            msg = json.loads(line)
            if msg.get("id") == self.i:
                if "error" in msg:
                    raise RuntimeError(msg["error"])
                return msg["result"]

    def close(self):
        self.p.stdin.close()
        self.p.wait(timeout=5)


CORE_TOOLS = {"sisyphus_new", "sisyphus_update", "sisyphus_show", "sisyphus_list", "sisyphus_search",
              "sisyphus_parent", "sisyphus_depends_on", "sisyphus_graph"}


def ollama_tools(mcp_tools, subset, schema_mode="raw"):
    out = []
    for t in mcp_tools:
        if subset == "core" and t["name"] not in CORE_TOOLS:
            continue
        schema = json.loads(json.dumps(t.get("inputSchema") or {"type": "object", "properties": {}}))
        schema.pop("$schema", None)
        schema.pop("additionalProperties", None)
        if schema_mode == "nodir":  # hide the optional repo dir; the server already runs in the repo
            schema.get("properties", {}).pop("dir", None)
        out.append({"type": "function", "function": {"name": t["name"], "description": t.get("description", ""),
                                                      "parameters": schema}})
    return out


# ---------------------------------------------------------------- prompts

PROMPTS = {
    "none": None,
    "basic": "You manage issues in this repository with the sisyphus tools. Use the tools to answer "
             "questions and to make changes. Do not guess: look issues up with the tools. "
             "When the work is done, reply with a short answer.",
    "detailed": (
        "You are a task manager for this repository. Issues are stored by the sisyphus tools; call them, "
        "never invent issue data.\n"
        "Facts:\n"
        "- An issue name is 2-6 lowercase words in kebab-case, for example fix-login-bug.\n"
        "- States: open, in-progress, closed. Priority: critical, high, medium, low. "
        "Effort: small, medium, large, x-large.\n"
        "- Close with resolution completed (done) or abandoned (will not do).\n"
        "- To start work: sisyphus_update with state in-progress, plus owner and bookmark if given.\n"
        "- To change priority/effort/tags without a state change, pass the current state to sisyphus_update.\n"
        "- A sub-issue has a parent (sisyphus_parent). A dependency (sisyphus_depends_on) means the "
        "first issue cannot start until the blocking issue closes.\n"
        "- To find issues use sisyphus_list (filters) or sisyphus_search (text). Use sisyphus_show for one issue.\n"
        "Steps: 1) look up what you need, 2) make each change with one tool call, 3) reply with a short answer "
        "that names the issues."),
}


# ---------------------------------------------------------------- run

def chat(host, body, timeout=1800):
    req = urllib.request.Request(f"{host}/api/chat", data=json.dumps(body).encode(),
                                 headers={"Content-Type": "application/json"})
    with urllib.request.urlopen(req, timeout=timeout) as r:
        return json.loads(r.read())


def exec_call(mcp, names, fn, a, args):
    if args.schema == "nodir":
        a.pop("dir", None)
    c = {"tool": fn, "args": a}
    if fn not in names:
        c["ok"], c["out"], c["unknown_tool"] = False, f"Error: unknown tool {fn}", True
    else:
        try:
            res = mcp.call("tools/call", {"name": fn, "arguments": a})
            c["out"] = "\n".join(x.get("text", "") for x in res.get("content", []))
            c["ok"] = not res.get("isError", False)
        except Exception as e:  # protocol-level error, e.g. bad argument types
            c["ok"], c["out"] = False, f"Error: {e}"
    return c


_CLIENT = None


def anthropic_loop(t, args, tools, mcp, rec):
    """Manual Messages API tool loop. Same tools (converted), prompts, and turn limit as the Ollama loop.
    Thinking is left at each model's default (adaptive); effort is set explicitly so models compare fairly.
    No refusal fallbacks: a fallback would hide which model produced the result."""
    global _CLIENT
    import anthropic
    if _CLIENT is None:
        # The sandbox proxy injects the real credential; the SDK only needs a placeholder to build the request.
        _CLIENT = anthropic.Anthropic(api_key=os.environ.get("ANTHROPIC_API_KEY", "proxy-managed"), max_retries=6)
    atools = [{"name": x["function"]["name"], "description": x["function"]["description"],
               "input_schema": x["function"]["parameters"]} for x in tools]
    names = {x["name"] for x in atools}
    kw = {"model": args.model, "max_tokens": 16000, "tools": atools,
          "output_config": {"effort": args.effort}, "cache_control": {"type": "ephemeral"}}
    if PROMPTS[args.prompt]:
        kw["system"] = PROMPTS[args.prompt]
    msgs = [{"role": "user", "content": t["prompt"]}]
    rec.update({"input_tokens": 0, "cache_read_tokens": 0, "cache_write_tokens": 0, "api_s": 0.0})
    for _ in range(args.max_turns):
        a0 = time.time()
        r = _CLIENT.messages.create(messages=msgs, **kw)
        rec["api_s"] += time.time() - a0
        rec["turns"] += 1
        u = r.usage
        rec["input_tokens"] += u.input_tokens or 0
        rec["cache_read_tokens"] += u.cache_read_input_tokens or 0
        rec["cache_write_tokens"] += u.cache_creation_input_tokens or 0
        rec["eval_tokens"] += u.output_tokens or 0
        rec["prompt_tokens"] += (u.input_tokens or 0) + (u.cache_read_input_tokens or 0) + (u.cache_creation_input_tokens or 0)
        msgs.append({"role": "assistant", "content": r.content})  # append-only: keeps thinking blocks valid
        if r.stop_reason == "refusal":
            rec["error"] = f"refusal: {getattr(r.stop_details, 'category', None)}"
            return
        uses = [b for b in r.content if b.type == "tool_use"]
        if not uses:
            rec["final"] = "".join(b.text for b in r.content if b.type == "text")
            return
        results = []
        for b in uses:
            c = exec_call(mcp, names, b.name, dict(b.input or {}), args)
            rec["calls"].append(c)
            results.append({"type": "tool_result", "tool_use_id": b.id, "content": c["out"][:6000] or "(no output)",
                            "is_error": not c["ok"]})
        msgs.append({"role": "user", "content": results})
    rec["error"] = "max_turns"


def run_task(t, args, tool_defs_cache):
    repo = make_repo()
    before = snapshot(repo)
    mcp = MCP(repo)
    if "tools" not in tool_defs_cache:
        tool_defs_cache["tools"] = mcp.call("tools/list", {})["tools"]
    tools = ollama_tools(tool_defs_cache["tools"], args.tools, args.schema)
    names = {t_["function"]["name"] for t_ in tools}
    msgs = []
    if PROMPTS[args.prompt]:
        msgs.append({"role": "system", "content": PROMPTS[args.prompt]})
    msgs.append({"role": "user", "content": t["prompt"]})
    rec = {"task": t["id"], "kind": t["kind"], "calls": [], "turns": 0, "prompt_tokens": 0, "eval_tokens": 0,
           "eval_s": 0.0, "prompt_s": 0.0, "load_s": 0.0, "error": None, "final": ""}
    t0 = time.time()
    if getattr(args, "backend", "ollama") == "anthropic":
        try:
            anthropic_loop(t, args, tools, mcp, rec)
        except Exception as e:
            rec["error"] = f"{type(e).__name__}: {e}"
        return finish_task(t, rec, t0, mcp, repo, before)
    try:
        for _ in range(args.max_turns):
            body = {"model": args.model, "messages": msgs, "tools": tools, "stream": False, "keep_alive": "30m",
                    "options": {"temperature": args.temperature, "num_ctx": args.num_ctx, "seed": args.seed}}
            if args.num_thread:  # 0 = let the server decide (GPU runs)
                body["options"]["num_thread"] = args.num_thread
            if args.think != "default":
                body["think"] = args.think == "on"
            r = chat(args.host, body)
            rec["turns"] += 1
            rec["prompt_tokens"] += r.get("prompt_eval_count", 0)
            rec["eval_tokens"] += r.get("eval_count", 0)
            rec["eval_s"] += r.get("eval_duration", 0) / 1e9
            rec["prompt_s"] += r.get("prompt_eval_duration", 0) / 1e9
            rec["load_s"] += r.get("load_duration", 0) / 1e9
            m = r["message"]
            msgs.append({k: v for k, v in m.items() if k in ("role", "content", "tool_calls", "thinking")})
            tcs = m.get("tool_calls") or []
            if not tcs:
                rec["final"] = m.get("content", "")
                break
            for tc in tcs:
                fn = tc["function"]["name"]
                a = tc["function"].get("arguments") or {}
                if isinstance(a, str):
                    try:
                        a = json.loads(a)
                    except Exception:
                        a = {}
                if args.schema == "nodir":
                    a.pop("dir", None)
                c = {"tool": fn, "args": a}
                if fn not in names:
                    c["ok"], c["out"], c["unknown_tool"] = False, f"Error: unknown tool {fn}", True
                else:
                    try:
                        res = mcp.call("tools/call", {"name": fn, "arguments": a})
                        c["out"] = "\n".join(x.get("text", "") for x in res.get("content", []))
                        c["ok"] = not res.get("isError", False)
                    except Exception as e:  # protocol-level error, e.g. bad argument types
                        c["ok"], c["out"] = False, f"Error: {e}"
                rec["calls"].append(c)
                msgs.append({"role": "tool", "tool_name": fn, "content": c["out"][:6000]})
        else:
            rec["error"] = "max_turns"
    except Exception as e:
        rec["error"] = f"{type(e).__name__}: {e}"
    return finish_task(t, rec, t0, mcp, repo, before)


def finish_task(t, rec, t0, mcp, repo, before):
    rec["wall_s"] = round(time.time() - t0, 2)
    mcp.close()
    after = snapshot(repo)
    checks = t["check"](before, after, rec["final"] or "")
    rec["checks"] = [{"ok": bool(ok), "what": w} for ok, w in checks]
    rec["score"] = sum(c["ok"] for c in rec["checks"]) / len(rec["checks"])
    rec["pass"] = all(c["ok"] for c in rec["checks"]) and rec["error"] is None
    rec["text_tool_call"] = bool(re.search(r"sisyphus_\w+|\"name\"\s*:", rec["final"] or "")) and not rec["calls"]
    shutil.rmtree(repo, ignore_errors=True)
    return rec


ORACLE = {
    "t01": ([], "fix-login-timeout, auth-epic, migrate-database"),
    "t02": ([], "fix-login-timeout"),
    "t03": ([], "alice, ai/docs-theme"),
    "t04": ([], "fix-login-timeout and add-password-reset"),
    "t05": ([], "No, there is none."),
    "t06": ([["new", "add-login-rate-limiting", "--title", "Add rate limiting to the login API", "--priority", "high",
              "--effort", "small", "--tags", "auth,feature"]], "done"),
    "t07": ([["update", "fix-login-timeout", "in-progress", "--owner", "bob", "--bookmark", "ai/login-timeout"]], "ok"),
    "t08": ([["update", "update-docs-site", "closed"]], "ok"),
    "t09": ([["update", "add-password-reset", "closed", "--resolution", "abandoned"]], "ok"),
    "t10": ([["parent", "migrate-database", "auth-epic"]], "ok"),
    "t11": ([["depends-on", "add-password-reset", "migrate-database"]], "ok"),
    "t12": ([["update", "migrate-database", "open", "--priority", "medium"]], "ok"),
    "t13": ([["new", "audit-session-tokens", "--title", "Audit session tokens", "--priority", "medium",
              "--parent", "auth-epic", "--depends-on", "fix-login-timeout"]], "ok"),
    "t14": ([["update", "migrate-database", "in-progress", "--owner", "carol", "--bookmark", "ai/db"]],
            "Migrate to Postgres 17"),
}


def run_oracle():
    (SCRATCH / "runs-tmp").mkdir(exist_ok=True)
    bad = 0
    for t in TASKS:
        repo = make_repo()
        b = snapshot(repo)
        cmds, ans = ORACLE[t["id"]]
        for c in cmds:
            sis(repo, *c)
        res = t["check"](b, snapshot(repo), ans)
        fails = [w for ok, w in res if not ok]
        bad += bool(fails)
        print(t["id"], "OK" if not fails else f"FAIL {fails}")
        shutil.rmtree(repo, ignore_errors=True)
    sys.exit(1 if bad else 0)


def agent_prepare(path, repeats):
    """Seed one repo per task run for an external agent (for example a Claude Code subagent)."""
    base = SCRATCH / "agent-runs"
    base.mkdir(parents=True, exist_ok=True)
    runs = []
    for rep in range(repeats):
        for t in TASKS:
            repo = make_repo()
            dest = base / f"{t['id']}-r{rep}"
            shutil.rmtree(dest, ignore_errors=True)
            shutil.move(str(repo), dest)  # neutral path, away from the harness and its checks
            runs.append({"task": t["id"], "repeat": rep, "kind": t["kind"], "prompt": t["prompt"],
                         "repo": str(dest), "before": snapshot(dest)})
    Path(path).write_text(json.dumps(runs, indent=1, default=str))
    print(f"prepared {len(runs)} runs in {base}")


def agent_score(manifest, answers, out, label):
    """answers: {"t01-r0": {"final": "...", "wall_s": 12.3, "tokens": 1234, "calls": 3}, ...}"""
    runs = json.loads(Path(manifest).read_text())
    ans = json.loads(Path(answers).read_text())
    by_id = {t["id"]: t for t in TASKS}
    recs = []
    for r in runs:
        key = f"{r['task']}-r{r['repeat']}"
        a = ans.get(key, {})
        before = {k: {kk: (str(vv) if kk in ("created", "closed") and vv is not None else vv) for kk, vv in v.items()}
                  for k, v in r["before"].items()}
        after = {k: {kk: (str(vv) if kk in ("created", "closed") and vv is not None else vv) for kk, vv in v.items()}
                 for k, v in snapshot(r["repo"]).items()}
        checks = by_id[r["task"]]["check"](before, after, a.get("final", ""))
        rec = {"task": r["task"], "kind": r["kind"], "repeat": r["repeat"], "final": a.get("final", ""),
               "wall_s": a.get("wall_s"), "agent_tokens": a.get("tokens"), "tool_uses": a.get("calls"),
               "checks": [{"ok": bool(ok), "what": w} for ok, w in checks], "error": None if a else "missing answer"}
        rec["score"] = sum(c["ok"] for c in rec["checks"]) / len(rec["checks"])
        rec["pass"] = all(c["ok"] for c in rec["checks"]) and bool(a)
        recs.append(rec)
        print(key, "PASS" if rec["pass"] else "fail", "; ".join(c["what"] for c in rec["checks"] if not c["ok"]))
    n = len(recs)
    walls = [r["wall_s"] for r in recs if r["wall_s"]]
    summary = {"model": label, "label": label, "config": {"backend": "claude-code-subagent", "interface": "sisyphus CLI via Bash"},
               "pass_rate": sum(r["pass"] for r in recs) / n, "mean_score": sum(r["score"] for r in recs) / n,
               "by_kind": {k: sum(r["pass"] for r in recs if r["kind"] == k) / max(1, sum(r["kind"] == k for r in recs))
                           for k in ("read", "write", "multi")},
               "by_task": {t["id"]: sum(r["pass"] for r in recs if r["task"] == t["id"]) /
                           max(1, sum(r["task"] == t["id"] for r in recs)) for t in TASKS},
               "mean_wall_s": round(sum(walls) / len(walls), 1) if walls else None, "runs": recs}
    Path(out).write_text(json.dumps(summary, indent=1))
    print(json.dumps({k: v for k, v in summary.items() if k != "runs"}, indent=1))


def model_info(host, model):
    try:
        req = urllib.request.Request(f"{host}/api/show", data=json.dumps({"model": model}).encode())
        s = json.loads(urllib.request.urlopen(req).read())
        d = s.get("details", {})
        return {"capabilities": s.get("capabilities"), "parameter_size": d.get("parameter_size"),
                "quantization": d.get("quantization_level"), "family": d.get("family")}
    except Exception as e:
        return {"error": str(e)}


def loaded_size(host, model):
    try:
        ps = json.loads(urllib.request.urlopen(f"{host}/api/ps").read())
        for m in ps.get("models", []):
            if m["name"] == model or m["model"] == model:
                return {"size": m.get("size"), "size_vram": m.get("size_vram")}
    except Exception:
        pass
    return None


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--model", required=True)
    ap.add_argument("--backend", choices=["ollama", "anthropic"], default="ollama")
    ap.add_argument("--effort", choices=["low", "medium", "high", "xhigh", "max"], default="medium",
                    help="anthropic backend only")
    ap.add_argument("--host", default=os.environ.get("OLLAMA_HOST_URL", "http://127.0.0.1:11434"))
    ap.add_argument("--think", choices=["on", "off", "default"], default="off")
    ap.add_argument("--prompt", choices=list(PROMPTS), default="basic")
    ap.add_argument("--tools", choices=["all", "core"], default="all")
    ap.add_argument("--schema", choices=["raw", "nodir"], default="raw")
    ap.add_argument("--oracle", action="store_true", help="check the checkers: run the known-correct commands")
    ap.add_argument("--temperature", type=float, default=0.2)
    ap.add_argument("--num-ctx", type=int, default=8192)
    ap.add_argument("--num-thread", type=int, default=16)
    ap.add_argument("--max-turns", type=int, default=8)
    ap.add_argument("--repeats", type=int, default=2)
    ap.add_argument("--seed", type=int, default=7)
    ap.add_argument("--duty", type=float, default=1.0, help="max busy fraction; <1 adds rest between tasks")
    ap.add_argument("--tasks", default="")
    ap.add_argument("--out", default=str(SCRATCH / "results"))
    ap.add_argument("--agent-prepare", metavar="MANIFEST")
    ap.add_argument("--agent-score", nargs=3, metavar=("MANIFEST", "ANSWERS", "OUT"))
    args = ap.parse_args()
    if args.agent_prepare:
        (SCRATCH / "runs-tmp").mkdir(exist_ok=True)
        return agent_prepare(args.agent_prepare, args.repeats)
    if args.agent_score:
        return agent_score(*args.agent_score, args.model)
    if args.oracle:
        run_oracle()
    (SCRATCH / "runs-tmp").mkdir(exist_ok=True)
    sel = [t for t in TASKS if not args.tasks or t["id"] in args.tasks.split(",")]
    cfg = {k: v for k, v in vars(args).items() if k not in ("out", "tasks", "host", "duty")}
    think_lbl = f"effort-{args.effort}" if args.backend == "anthropic" else f"think-{args.think}"
    label = f"{args.model.replace(':', '_').replace('/', '_')}__{think_lbl}__prompt-{args.prompt}__tools-{args.tools}__schema-{args.schema}"
    out = Path(args.out) / f"{label}.json"
    out.parent.mkdir(parents=True, exist_ok=True)
    runs = []
    for rep in range(args.repeats):
        for t in sel:
            r = run_task(t, argparse.Namespace(**{**vars(args), "seed": args.seed + rep}), {})
            while (SCRATCH / "PAUSE").exists():  # manual pause switch for the shared workstation
                time.sleep(15)
            if args.duty < 1:  # rest between tasks so a shared GPU is busy at most `duty` of the time
                time.sleep(r["wall_s"] * (1 / args.duty - 1))
            r["repeat"] = rep
            if r["error"] and r["error"].split(":")[0] in ("HTTPError", "URLError", "ConnectionRefusedError", "RemoteDisconnected"):
                print(f"[{args.model}] server error on {t['id']}: {r['error']} -- stopping this config", flush=True)
                sys.exit(6)
            runs.append(r)
            mark = "PASS" if r["pass"] else "fail"
            fails = "; ".join(c["what"] for c in r["checks"] if not c["ok"])
            print(f"[{args.model}] rep{rep} {t['id']} {mark} score={r['score']:.2f} calls={len(r['calls'])} "
                  f"wall={r['wall_s']}s {r['error'] or ''} {fails}", flush=True)
    n = len(runs)
    ev_tok = sum(r["eval_tokens"] for r in runs)
    ev_s = sum(r["eval_s"] for r in runs)
    calls = [c for r in runs for c in r["calls"]]
    summary = {
        "model": args.model, "label": label, "config": cfg, "info": model_info(args.host, args.model) if args.backend == "ollama" else {"backend": "anthropic"},
        "loaded_bytes": loaded_size(args.host, args.model) if args.backend == "ollama" else None,
        "pass_rate": sum(r["pass"] for r in runs) / n,
        "mean_score": sum(r["score"] for r in runs) / n,
        "by_kind": {k: sum(r["pass"] for r in runs if r["kind"] == k) / max(1, sum(r["kind"] == k for r in runs))
                    for k in ("read", "write", "multi")},
        "by_task": {t["id"]: sum(r["pass"] for r in runs if r["task"] == t["id"]) / args.repeats for t in sel},
        "tool_calls": len(calls), "tool_call_errors": sum(not c["ok"] for c in calls),
        "unknown_tool_calls": sum(bool(c.get("unknown_tool")) for c in calls),
        "no_tool_runs": sum(not r["calls"] for r in runs),
        "text_tool_call_runs": sum(r["text_tool_call"] for r in runs),
        "errors": sum(r["error"] is not None for r in runs),
        "gen_tok_per_s": round(ev_tok / ev_s, 2) if ev_s else None,
        "prompt_tok_per_s": round(sum(r["prompt_tokens"] for r in runs) / max(1e-9, sum(r["prompt_s"] for r in runs)), 1),
        "mean_wall_s": round(sum(r["wall_s"] for r in runs) / n, 1),
        "mean_eval_tokens": round(ev_tok / n, 1),
        "runs": runs,
    }
    out.write_text(json.dumps(summary, indent=1))
    print(json.dumps({k: v for k, v in summary.items() if k != "runs"}, indent=1))
    print(f"wrote {out}")


if __name__ == "__main__":
    main()
