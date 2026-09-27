// RunInit writes .gitagger.yml with commented defaults.
package cmd

import (
	"fmt"
	"os"

	"github.com/CHE3MZ/gitagger/internal/config"
	"github.com/CHE3MZ/gitagger/internal/style"
)

// RunInit creates .gitagger.yml in dir unless it exists without force.
func RunInit(dir string, force bool) error {
	existing := config.Find(dir)
	if existing != "" && !force {
		return Exists("config already exists at %s (use --force to overwrite)", existing)
	}
	path := dir + string(os.PathSeparator) + ".gitagger.yml"
	if err := os.WriteFile(path, []byte(config.DefaultFileContent()), 0o600); err != nil {
		return Generic("couldn't write %s (%v)", path, err)
	}
	fmt.Printf("%s %s\n", style.Green("wrote"), style.White(".gitagger.yml"))
	fmt.Println(style.Dim("tweak it if you like, then just run `gitagger`"))
	return nil
}
