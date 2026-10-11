package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

// linkChange is one wikilink that a rename rewrites.
type linkChange struct {
	File string `json:"file"`
	Line int    `json:"line"`
	From string `json:"from"`
	To   string `json:"to"`
}

// renamePlan is what a rename of an issue changes: the file, and every link to it.
type renamePlan struct {
	From  string       `json:"from"`
	To    string       `json:"to"`
	Links []linkChange `json:"links"`
	// contents holds the new text of each file whose links change, by absolute path. The text of the
	// renamed issue is under its old path.
	contents map[string]string
	name     string
	newName  string
	issue    issueFile
}

// planRename checks a rename of the issue reference to newName, and finds every link to change. It
// changes nothing.
func planRename(root, reference, newName string, warnings io.Writer) (renamePlan, error) {
	newName = strings.Trim(trimMarkdownExtension(strings.ReplaceAll(newName, `\`, "/")), "/")
	x := loadIndex(root)
	name, issue, _, err := x.load(reference)
	if err != nil {
		return renamePlan{}, err
	}
	if newName == name {
		return renamePlan{}, fmt.Errorf("Issue '%s' already has the name '%s'.", name, newName)
	}
	if !validName(newName) {
		return renamePlan{}, fmt.Errorf("Invalid issue name '%s'. Use 2-6 lowercase words in kebab-case, for example fix-login-bug, optionally below kebab-case directories, for example web/fix-login-bug.", newName)
	}
	if existing := x.exactOf(newName); len(existing) > 0 {
		return renamePlan{}, fmt.Errorf("Issue '%s' already exists: %s.", newName, relative(root, existing[0].path))
	}
	if c, ok := openClaims(root).claimedElsewhere(name); ok {
		return renamePlan{}, fmt.Errorf("Issue '%s' is in progress in another workspace, %s, owned by %s. Its work still uses the old name, so rename the issue after that work merges.", name, openClaims(root).display(c.Workspace), describeOwner(c.Owner))
	}
	destination := issuePath(root, issue.state, newName)
	plan := renamePlan{
		From:     relative(root, issue.path),
		To:       relative(root, destination),
		Links:    []linkChange{},
		contents: map[string]string{},
		name:     name,
		newName:  newName,
		issue:    issue,
	}

	candidates := repoCandidates(root)
	// The new file name is unique when no other Markdown file has it. Then a link with only the file
	// name still finds the issue after the rename.
	newBase := strings.ToLower(baseName(newName))
	var clashes []string
	for _, c := range candidates.list {
		if c.file != issue.path && hasSuffixFold(c.relative, ".md") && strings.ToLower(trimMarkdownExtension(filepath.Base(c.relative))) == newBase {
			clashes = append(clashes, c.relative)
		}
	}
	if len(clashes) > 0 {
		warn(warnings, fmt.Sprintf("Other files also have the name '%s': %s. Links to the issue get its full name, [[%s]].", baseName(newName), strings.Join(clashes, ", "), newName))
	}
	for _, c := range candidates.list {
		if !hasSuffixFold(c.relative, ".md") {
			continue
		}
		trace(c.file)
		content, err := os.ReadFile(c.file)
		if err != nil {
			continue
		}
		changed, changes := rewriteLinks(string(content), func(target string, inFrontmatter bool) (string, bool) {
			matches := matchTarget(candidates, target, false)
			if len(matches) != 1 || matches[0].file != issue.path {
				return "", false
			}
			switch {
			case inFrontmatter || strings.Contains(target, "/") || len(clashes) > 0:
				// sisyphus writes the full name in the frontmatter, so that a later file with the same
				// name does not make the link ambiguous (see commands#Naming an issue).
				replacement := newName
				if hasSuffixFold(target, ".md") {
					replacement += ".md"
				}
				return replacement, replacement != target
			default:
				replacement := baseName(newName)
				if hasSuffixFold(target, ".md") {
					replacement += ".md"
				}
				return replacement, replacement != target
			}
		})
		if len(changes) == 0 {
			continue
		}
		plan.contents[c.file] = changed
		for _, change := range changes {
			change.File = c.relative
			plan.Links = append(plan.Links, change)
		}
	}
	return plan, nil
}

// rewriteLinks returns content with each wikilink target that replace changes, and the changes. It
// skips code blocks and inline code, as Obsidian does, and keeps the heading, the text after "|", and
// the line endings of each link.
func rewriteLinks(content string, replace func(target string, inFrontmatter bool) (string, bool)) (string, []linkChange) {
	lines := strings.SplitAfter(content, "\n")
	var changes []linkChange
	inFence, inFrontmatter := false, false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 && strings.TrimPrefix(trimmed, utf8BOM) == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && trimmed == "---" {
			inFrontmatter = false
			continue
		}
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		code := inlineCodePattern.FindAllStringIndex(line, -1)
		var rebuilt strings.Builder
		last := 0
		for _, match := range wikilinkPattern.FindAllStringSubmatchIndex(line, -1) {
			if slices.ContainsFunc(code, func(span []int) bool { return match[0] >= span[0] && match[0] < span[1] }) {
				continue
			}
			inner := line[match[2]:match[3]]
			end := strings.IndexAny(inner, "#|")
			if end < 0 {
				end = len(inner)
			}
			target := strings.TrimSpace(inner[:end])
			if target == "" {
				continue // [[#Heading]] points into its own file
			}
			replacement, ok := replace(target, inFrontmatter)
			if !ok {
				continue
			}
			link := "[[" + replacement + inner[end:] + "]]"
			changes = append(changes, linkChange{Line: i + 1, From: line[match[0]:match[1]], To: link})
			rebuilt.WriteString(line[last:match[0]])
			rebuilt.WriteString(link)
			last = match[1]
		}
		if last > 0 {
			rebuilt.WriteString(line[last:])
			lines[i] = rebuilt.String()
		}
	}
	return strings.Join(lines, ""), changes
}

// renameIssue renames the issue reference to newName: it moves the file in its state directory,
// rewrites every link to it, and moves its claim. The caller holds the lock.
func renameIssue(root, reference, newName string, warnings io.Writer) (string, error) {
	plan, err := planRename(root, reference, newName, warnings)
	if err != nil {
		return "", err
	}
	if err := applyRename(root, plan); err != nil {
		return "", err
	}
	return plan.To, nil
}

func applyRename(root string, plan renamePlan) error {
	// The renamed issue first, so that its links are right when it moves.
	if content, ok := plan.contents[plan.issue.path]; ok {
		if err := writeFileAtomically(plan.issue.path, []byte(content)); err != nil {
			return err
		}
	}
	destination := filepath.Join(root, filepath.FromSlash(plan.To))
	if err := moveIssueFile(root, plan.issue.path, destination); err != nil {
		return err
	}
	var failed []string
	for file, content := range plan.contents {
		if file == plan.issue.path {
			continue
		}
		if err := writeFileAtomically(file, []byte(content)); err != nil {
			failed = append(failed, fmt.Sprintf("%s (%v)", relative(root, file), err))
		}
	}
	claims := openClaims(root)
	if c, ok := claims.get(plan.name); ok && claims.isHere(c) {
		c.Name = plan.newName
		claims.remove(plan.name)
		if err := claims.put(c); err != nil {
			failed = append(failed, "the claim ("+err.Error()+")")
		}
	}
	if len(failed) > 0 {
		slices.Sort(failed)
		return errors.New("The issue moved to " + plan.To + ", but these files could not be changed: " + strings.Join(failed, ", ") + ". Run sisyphus lint, and correct their links.")
	}
	return nil
}

func renameCommand(findRoot func() (string, error)) *cobra.Command {
	var dryRun bool
	cmd := &cobra.Command{
		Use:   "rename <name> <new-name>",
		Short: "Rename an issue or move it to a directory, and update its links.",
		Long: "Rename an issue, or move it to another subdirectory of its state directory, and change every " +
			"wikilink to it in the repo, also in the frontmatter of other issues.\n\n" +
			"<new-name> is a full name: 2-6 lowercase words in kebab-case, optionally below kebab-case " +
			"directories, for example web/fix-login-bug. The issue stays in its state, and its title does not " +
			"change. A link that has only the file name stays so while no other file has that name; other links " +
			"and the frontmatter get the full name. --dry-run prints what would change, and changes nothing.\n\n" +
			"Links in other branches and workspaces, and links outside the repo, do not change.",
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !dryRun {
				return printPath(cmd, findRoot, func(root string, warnings io.Writer) (string, error) {
					return renameIssue(root, args[0], args[1], warnings)
				})
			}
			root, err := findRoot()
			if err != nil {
				return err
			}
			plan, err := planRename(root, args[0], args[1], cmd.ErrOrStderr())
			if err != nil {
				return err
			}
			slices.SortFunc(plan.Links, func(a, b linkChange) int {
				if c := strings.Compare(a.File, b.File); c != 0 {
					return c
				}
				return a.Line - b.Line
			})
			if asJSON, _ := cmd.Flags().GetBool("json"); asJSON {
				encoder := json.NewEncoder(cmd.OutOrStdout())
				encoder.SetEscapeHTML(false)
				encoder.SetIndent("", "  ")
				return encoder.Encode(plan)
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "%s -> %s\n", plan.From, plan.To)
			for _, change := range plan.Links {
				fmt.Fprintf(out, "%s:%d: %s -> %s\n", change.File, change.Line, change.From, change.To)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Print the move and each link that would change, and change nothing.")
	return cmd
}
