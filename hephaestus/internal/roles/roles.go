// Package roles reads the role files of hephaestus: one Markdown file for each role,
// roles/<role>.md, or roles/<family>/<variant>/role.md for a role of a family, for example
// roles/formatter/golang/role.md for the role formatter/golang. A role file has YAML frontmatter that hephaestus reads and sections that describe the role to
// people and, for an AI role, to the model. design/role-file.md describes the format.
package roles

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
)

// Kinds of role.
const (
	AI   = "ai"
	Tool = "tool"
)

var (
	// Triggers are the events that start a role (Q5 of the agent delegation design).
	Triggers = []string{"request", "close", "schedule"}
	// Sections are the sections that every role file has.
	Sections = []string{"Charter", "Inputs", "Outputs", "Tools", "Ask a human"}

	namePattern    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	rolePattern    = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*(/[a-z0-9]+(-[a-z0-9]+)*)*$`)
	headingPattern = regexp.MustCompile(`^##\s+(.+?)\s*$`)
)

// Role is one role file.
type Role struct {
	// Name is the path of the role in the roles directory: merge-queue, or formatter/golang.
	Name              string   `json:"name"`
	Path              string   `json:"path"`
	Title             string   `json:"title"`
	Kind              string   `json:"kind"`
	Command           string   `json:"command,omitempty"`
	Triggers          []string `json:"triggers"`
	OnFailure         string   `json:"on-failure"`
	CanCallReferendum bool     `json:"can-call-referendum"`
	// Listens holds the kinds of broadcast that the role reacts to, for example bookmark-ready. Its
	// workers also get every message to them and to the role.
	Listens []string `json:"listens,omitempty"`
	// AllowedTools are the Claude Code tools that an AI role may use, for example [Bash, Read]. Empty:
	// the tools of the session.
	AllowedTools []string `json:"allowed-tools,omitempty"`
	// Model is the model of an AI role, for example sonnet. Empty: the model of the session.
	Model string `json:"model,omitempty"`
	// Charter is the text of the Charter section, for a summary.
	Charter  string `json:"charter"`
	sections map[string]string
}

// Dir returns the roles directory of the repo at root: hephaestus/roles, or roles when root is the
// hephaestus project itself.
func Dir(root string) string {
	for _, dir := range []string{filepath.Join(root, "hephaestus", "roles"), filepath.Join(root, "roles")} {
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return filepath.Join(root, "hephaestus", "roles")
}

// RoleFile is the name of the role file in the directory of a role of a family.
const RoleFile = "role.md"

// Load reads every role file of dir, sorted by name: each <role>.md of dir, and each role.md below
// it. Other files below dir, for example a README or a script, belong to the roles and are not role
// files. It returns the roles that it could read, and one error that lists every problem of every
// file.
func Load(dir string) ([]Role, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}
		if filepath.Dir(path) == filepath.Clean(dir) || (entry.Name() == RoleFile && filepath.Dir(path) != filepath.Clean(dir)) {
			files = append(files, path)
		}
		return nil
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, err
	}
	var roles []Role
	var problems []string
	for _, file := range files {
		role, fileProblems := read(dir, file)
		if len(fileProblems) > 0 {
			rel, _ := filepath.Rel(dir, file)
			for _, p := range fileProblems {
				problems = append(problems, fmt.Sprintf("%s: %s", filepath.ToSlash(rel), p))
			}
			continue
		}
		roles = append(roles, role)
	}
	names := map[string]bool{}
	for _, role := range roles {
		names[role.Name] = true
	}
	for _, role := range roles {
		if role.OnFailure != "needs-human" {
			target := slices.IndexFunc(roles, func(r Role) bool { return r.Name == role.OnFailure })
			switch {
			case target < 0:
				problems = append(problems, fmt.Sprintf("%s: on-failure names the role %q, which has no role file. Use needs-human or a role.", role.File(), role.OnFailure))
			case roles[target].Kind != AI:
				problems = append(problems, fmt.Sprintf("%s: on-failure names %q, which is not an AI role. Work that a tool cannot do goes to an AI role or a human.", role.File(), role.OnFailure))
			}
		}
	}
	slices.SortFunc(roles, func(a, b Role) int { return strings.Compare(a.Name, b.Name) })
	if len(problems) > 0 {
		return roles, errors.New(strings.Join(problems, "\n"))
	}
	return roles, nil
}

// Find returns the role named name in dir: dir/<name>.md, or dir/<name>/role.md.
func Find(dir, name string) (Role, error) {
	file := ""
	if rolePattern.MatchString(name) {
		for _, candidate := range []string{filepath.Join(dir, name+".md"), filepath.Join(dir, filepath.FromSlash(name), RoleFile)} {
			if strings.Contains(name, "/") && !strings.HasSuffix(candidate, RoleFile) {
				continue // a file below dir is a role file only when its name is role.md
			}
			if _, err := os.Stat(candidate); err == nil {
				file = candidate
				break
			}
		}
	}
	if file == "" {
		known, _ := Load(dir)
		var names []string
		for _, r := range known {
			names = append(names, r.Name)
		}
		if len(names) == 0 {
			return Role{}, fmt.Errorf("No role %q: %s has no role files.", name, dir)
		}
		return Role{}, fmt.Errorf("No role %q in %s. The roles are: %s.", name, dir, strings.Join(names, ", "))
	}
	role, problems := read(dir, file)
	if len(problems) > 0 {
		return Role{}, fmt.Errorf("%s: %s", role.File(), strings.Join(problems, "; "))
	}
	return role, nil
}

// nameOf returns the name of the role of a role file of dir: its path without .md, or the path of
// its directory for a role.md.
func nameOf(dir, file string) string {
	rel, err := filepath.Rel(dir, file)
	if err != nil {
		rel = filepath.Base(file)
	}
	rel = filepath.ToSlash(rel)
	if strings.Contains(rel, "/") && path.Base(rel) == RoleFile {
		return path.Dir(rel)
	}
	return strings.TrimSuffix(rel, ".md")
}

// File returns the path of the role file from the roles directory, for a message: merge-queue.md,
// or formatter/golang/role.md.
func (r Role) File() string {
	if strings.Contains(r.Name, "/") {
		return r.Name + "/" + RoleFile
	}
	return r.Name + ".md"
}

// Flat returns the name of the role with "-" for "/", for a file name or a worker name:
// formatter-golang for formatter/golang.
func (r Role) Flat() string {
	return strings.ReplaceAll(r.Name, "/", "-")
}

// Dir returns the directory of the role file. A role of a family keeps the files of its
// implementation there, for example a script.
func (r Role) Dir() string {
	return filepath.Dir(r.Path)
}

// read reads one role file of dir, and returns its problems.
func read(dir, file string) (Role, []string) {
	data, err := os.ReadFile(file)
	if err != nil {
		return Role{}, []string{err.Error()}
	}
	text := strings.ReplaceAll(string(data), "\r\n", "\n")
	role := Role{Name: nameOf(dir, file), Path: file, OnFailure: "needs-human", sections: map[string]string{}}
	var problems []string
	front, body, ok := splitFrontmatter(text)
	if !ok {
		return role, []string{"the file must start with frontmatter between two --- lines"}
	}
	fields := map[string]string{}
	for _, line := range strings.Split(front, "\n") {
		line = strings.TrimSpace(stripComment(line))
		if line == "" {
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if !found {
			problems = append(problems, fmt.Sprintf("the frontmatter line %q is not key: value", line))
			continue
		}
		fields[strings.TrimSpace(key)] = unquote(strings.TrimSpace(value))
	}
	if fields["role"] != role.Name {
		problems = append(problems, fmt.Sprintf("role is %q, but the path of the file gives %q: they must be the same", fields["role"], role.Name))
	}
	if !rolePattern.MatchString(role.Name) {
		problems = append(problems, "the role name must be lowercase words in kebab-case, for example merge-queue, or a path of them for a role of a family, for example formatter/golang")
	}
	role.Kind = fields["kind"]
	if role.Kind != AI && role.Kind != Tool {
		problems = append(problems, fmt.Sprintf("kind is %q: use ai or tool", role.Kind))
	}
	role.Command = fields["command"]
	if role.Kind == Tool && role.Command == "" {
		problems = append(problems, "a tool role needs a command: the program that does its work")
	}
	if role.Kind == AI && role.Command != "" {
		problems = append(problems, "an AI role has no command: a model does its work, from its Prompt section")
	}
	for _, trigger := range strings.Split(strings.Trim(fields["trigger"], "[]"), ",") {
		if trigger = strings.TrimSpace(trigger); trigger == "" {
			continue
		}
		if !slices.Contains(Triggers, trigger) {
			problems = append(problems, fmt.Sprintf("trigger %q is not one of: %s", trigger, strings.Join(Triggers, ", ")))
		}
		role.Triggers = append(role.Triggers, trigger)
	}
	if len(role.Triggers) == 0 {
		problems = append(problems, "trigger is missing: "+strings.Join(Triggers, ", ")+", or a list of them")
	}
	if value := fields["on-failure"]; value != "" {
		role.OnFailure = value
	}
	for _, kind := range strings.Split(strings.Trim(fields["listens"], "[]"), ",") {
		if kind = strings.TrimSpace(kind); kind == "" {
			continue
		}
		if !namePattern.MatchString(kind) {
			problems = append(problems, fmt.Sprintf("listens has %q, which is not a kind: lowercase words, for example bookmark-ready", kind))
		}
		role.Listens = append(role.Listens, kind)
	}
	for _, tool := range strings.Split(strings.Trim(fields["allowed-tools"], "[]"), ",") {
		if tool = strings.TrimSpace(tool); tool != "" {
			role.AllowedTools = append(role.AllowedTools, tool)
		}
	}
	role.Model = fields["model"]
	if role.Kind == Tool && (len(role.AllowedTools) > 0 || role.Model != "") {
		problems = append(problems, "allowed-tools and model are for an AI role; a tool role runs its command")
	}
	switch fields["can-call-referendum"] {
	case "", "false":
	case "true":
		role.CanCallReferendum = true
	default:
		problems = append(problems, "can-call-referendum is true or false")
	}

	section := ""
	for _, line := range strings.Split(body, "\n") {
		if m := headingPattern.FindStringSubmatch(line); m != nil {
			section = m[1]
			role.sections[section] = ""
			continue
		}
		if strings.HasPrefix(line, "# ") && role.Title == "" {
			role.Title = strings.TrimSpace(strings.TrimPrefix(line, "# "))
			continue
		}
		if section != "" {
			role.sections[section] += line + "\n"
		}
	}
	if role.Title == "" {
		problems = append(problems, `the body needs a title: "# <Role name>"`)
	}
	required := slices.Clone(Sections)
	if role.Kind == AI {
		required = append(required, "Prompt")
	}
	for _, name := range required {
		if strings.TrimSpace(role.sections[name]) == "" {
			problems = append(problems, fmt.Sprintf("the section %q is missing or empty", "## "+name))
		}
	}
	role.Charter = firstParagraph(role.sections["Charter"])
	return role, problems
}

// Section returns the text of a section of the role file.
func (r Role) Section(name string) string {
	return strings.TrimSpace(r.sections[name])
}

func splitFrontmatter(text string) (front, body string, ok bool) {
	if !strings.HasPrefix(text, "---\n") {
		return "", "", false
	}
	end := strings.Index(text[4:], "\n---\n")
	if end < 0 {
		return "", "", false
	}
	return text[4 : 4+end], text[4+end+5:], true
}

// stripComment removes a "#" comment that is not inside quotes.
func stripComment(line string) string {
	inQuote := rune(0)
	for i, r := range line {
		switch {
		case inQuote != 0 && r == inQuote:
			inQuote = 0
		case inQuote == 0 && (r == '"' || r == '\''):
			inQuote = r
		case inQuote == 0 && r == '#' && (i == 0 || line[i-1] == ' '):
			return line[:i]
		}
	}
	return line
}

func firstParagraph(text string) string {
	text = strings.TrimSpace(text)
	if i := strings.Index(text, "\n\n"); i >= 0 {
		text = text[:i]
	}
	return strings.Join(strings.Fields(text), " ")
}

// unquote removes the quotes around a value that is quoted as a whole, for example "a b". A quote
// inside a value, for example in a command line, stays.
func unquote(value string) string {
	if len(value) >= 2 && (value[0] == '"' || value[0] == '\'') && value[len(value)-1] == value[0] {
		return value[1 : len(value)-1]
	}
	return value
}
