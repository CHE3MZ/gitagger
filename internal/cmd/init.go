// RunInit generates a .gitagger.yml file.
package cmd

import (
	"fmt"
	"os"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunInit writes .gitagger.yml in dir. Refuses to overwrite unless force.
// With verbose it prints the full path. Without comment it strips all
// comments from the template; with comment it keeps them.
func RunInit(dir string, force, verbose, comment bool) error {
	existing := config.Find(dir)
	if existing != "" && !force {
		return Exists("config already exists at %s (use -f to overwrite)", existing)
	}
	path := dir + string(os.PathSeparator) + ".gitagger.yml"
	content := config.CleanFileContent()
	if comment {
		content = config.DefaultFileContent()
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return Generic("couldn't write %s (%v)", path, err)
	}
	if verbose {
		fmt.Printf("%s %s\n", style.Green("wrote"), style.White(path))
	} else {
		fmt.Printf("%s %s\n", style.Green("wrote"), style.White(".gitagger.yml"))
	}
	return nil
}
