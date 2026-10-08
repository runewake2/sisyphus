# sisyphus slug

Turn any text, for example a title, into a valid ASCII issue name that no file in the repo has yet.

## Usage

```bash
sisyphus slug <text>
```

## Rules

1. Lowercase the text and spell its letters in ASCII. Accents go (`é` becomes `e`, `ñ` becomes
   `n`), and letters without an accent form get an ASCII spelling (`ß` becomes `ss`, `æ` becomes
   `ae`, `ø` becomes `o`, `ł` becomes `l`). The name is always ASCII.
2. Each run of ASCII letters and digits is a word. Every other character, for example a Cyrillic
   or CJK letter, an emoji, or punctuation, separates words and is dropped.
3. Drop the filler words `a`, `an`, `and`, `are`, `at`, `be`, `by`, `for`, `from`, `in`, `into`,
   `is`, `of`, `on`, `or`, `the`, `to`, `was`, `were`, and `with`. If fewer than 2 words would
   remain, keep them all.
4. Keep the first 6 words and drop the rest. Fewer than 2 words is an error: the text cannot make
   a valid name.
5. Join the words with hyphens.
6. If a Markdown file anywhere in the repo already has that name (ignoring case), add `-2`, `-3`,
   and so on until it does not. If the name already has 6 words, the last word gives way to the
   number, so the result always has 2-6 words.

`slug` only prints a name; it does not create anything. Two `slug` calls before a `new` can return
the same name.

## Output

The name, for example:

```bash
$ sisyphus slug "Crash when the cache is empty!"
crash-when-cache-empty
$ sisyphus slug "Build the issue index once per command"
build-issue-index-once-per-command
$ sisyphus slug "Café crème: fix the naïve parser"
cafe-creme-fix-naive-parser
$ sisyphus slug "Fix the login bug"     # fix-login-bug.md already exists
fix-login-bug-2
```

## Examples

```bash
name="$(sisyphus slug "$TITLE")" && sisyphus new "$name" --title "$TITLE"
```
