# sisyphus slug

Turn any text, for example a title, into a valid issue name that no file in the repo has yet.

## Usage

```bash
sisyphus slug <text>
```

## Rules

1. Lowercase the text and keep only runs of letters and digits, as words.
2. Keep the first 6 words. Fewer than 2 words is an error: the text cannot make a valid name.
3. Join the words with hyphens.
4. If a Markdown file anywhere in the repo already has that name (ignoring case), add `-2`, `-3`,
   and so on until it does not. If the name already has 6 words, the last word gives way to the
   number, so the result always has 2-6 words.

`slug` only prints a name; it does not create anything. Two `slug` calls before a `new` can return
the same name.

## Output

The name, for example:

```bash
$ sisyphus slug "Crash when the cache is empty!"
crash-when-the-cache-is-empty
$ sisyphus slug "Fix login bug"     # fix-login-bug.md already exists
fix-login-bug-2
```

## Examples

```bash
name="$(sisyphus slug "$TITLE")" && sisyphus new "$name" --title "$TITLE"
```
