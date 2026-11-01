package securitycontact

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"unicode/utf8"
)

var emailAddress = regexp.MustCompile(`(?i)[a-z0-9._%+-]+@[a-z0-9.-]+\.[a-z]{2,}`)

const approvedBackupEmail = "SuperMcDuck@proton.me"

func TestRepositoryEmailAddressesUseApprovedSecurityContact(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate repository root")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "../.."))
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", "licenses", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		normalized := filepath.ToSlash(relative)
		if normalized == "go.sum" || isThirdPartyLicenseText(normalized) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !utf8.Valid(data) {
			return nil
		}
		for _, match := range emailAddress.FindAllString(string(data), -1) {
			lower := strings.ToLower(match)
			if match == approvedBackupEmail || strings.HasSuffix(lower, "@users.noreply.github.com") || strings.HasSuffix(lower, "@noreply.github.com") {
				continue
			}
			t.Errorf("unapproved email address %q in %s", match, normalized)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func isThirdPartyLicenseText(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	return base == "third_party_notices.md" || base == "licenses.md"
}
