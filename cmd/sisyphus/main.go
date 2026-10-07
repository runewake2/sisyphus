// Command sisyphus manages the issues in issues/ and resolves wikilinks in the repo.
package main

import "os"

func main() {
	os.Exit(run(os.Args[1:], findRoot, os.Stdout, os.Stderr))
}
