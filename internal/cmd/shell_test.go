// Shell resolution tests: platform defaults and the Windows fallback chain.
package cmd

import (
	"errors"
	"testing"
)

func lookOnly(found ...string) func(string) (string, error) {
	return func(name string) (string, error) {
		for _, f := range found {
			if f == name {
				return "/fake/" + name, nil
			}
		}
		return "", errors.New("not found")
	}
}

func TestShellDefaultUnixIsSh(t *testing.T) {
	for _, goos := range []string{"linux", "darwin"} {
		bin, prefix, err := shellFor(goos, "", lookOnly())
		if err != nil || bin != "sh" || len(prefix) != 1 || prefix[0] != "-c" {
			t.Fatalf("%s default = %q %v %v", goos, bin, prefix, err)
		}
	}
}

func TestShellDefaultWindowsPrefersPwsh(t *testing.T) {
	bin, prefix, err := shellFor("windows", "", lookOnly("pwsh", "powershell", "cmd"))
	if err != nil || bin != "pwsh" || len(prefix) != 1 || prefix[0] != "-Command" {
		t.Fatalf("windows default = %q %v %v", bin, prefix, err)
	}
}

func TestShellDefaultWindowsFallsBackToWindowsPowerShell(t *testing.T) {
	bin, prefix, err := shellFor("windows", "", lookOnly("powershell", "cmd"))
	if err != nil || bin != "powershell" || len(prefix) != 1 || prefix[0] != "-Command" {
		t.Fatalf("windows fallback = %q %v %v", bin, prefix, err)
	}
}

func TestShellDefaultWindowsFallsBackToCmd(t *testing.T) {
	bin, prefix, err := shellFor("windows", "", lookOnly("cmd"))
	if err != nil || bin != "cmd" || len(prefix) != 1 || prefix[0] != "/C" {
		t.Fatalf("windows cmd fallback = %q %v %v", bin, prefix, err)
	}
}

func TestShellDefaultWindowsFallsBackToBash(t *testing.T) {
	bin, prefix, err := shellFor("windows", "", lookOnly("bash"))
	if err != nil || bin != "bash" || len(prefix) != 1 || prefix[0] != "-c" {
		t.Fatalf("windows bash fallback = %q %v %v", bin, prefix, err)
	}
}

func TestShellDefaultWindowsNoShell(t *testing.T) {
	if _, _, err := shellFor("windows", "", lookOnly()); err == nil {
		t.Fatalf("windows with no shell on PATH should fail")
	}
}

func TestShellPwshOnWindowsFallsBackToo(t *testing.T) {
	bin, _, err := shellFor("windows", "pwsh", lookOnly("powershell", "cmd"))
	if err != nil || bin != "powershell" {
		t.Fatalf("pwsh on windows = %q %v", bin, err)
	}
}

func TestShellPwshOnUnixStaysExact(t *testing.T) {
	bin, prefix, err := shellFor("linux", "pwsh", lookOnly())
	if err != nil || bin != "pwsh" || len(prefix) != 1 || prefix[0] != "-Command" {
		t.Fatalf("pwsh on linux = %q %v %v", bin, prefix, err)
	}
}

func TestShellExplicitShOnWindowsStaysExact(t *testing.T) {
	bin, prefix, err := shellFor("windows", "sh", lookOnly())
	if err != nil || bin != "sh" || len(prefix) != 1 || prefix[0] != "-c" {
		t.Fatalf("explicit sh on windows = %q %v %v", bin, prefix, err)
	}
}

func TestShellBatchNeedsWindows(t *testing.T) {
	if _, _, err := shellFor("linux", "batch", lookOnly()); err == nil {
		t.Fatalf("batch on linux should fail")
	}
}

func TestShellUnknownRejected(t *testing.T) {
	if _, _, err := shellFor("linux", "fish", lookOnly()); err == nil {
		t.Fatalf("unknown shell should fail")
	}
}
