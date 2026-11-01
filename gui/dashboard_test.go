//go:build !nogui

package gui

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2/test"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
	"github.com/NexzaDev/attackpath-visualizer/internal/service"
)

func TestEveryNavigationCategoryBuildsARealScreen(t *testing.T) {
	application := test.NewApp()
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err = (service.Service{Store: store}).LoadDemo(context.Background()); err != nil {
		t.Fatal(err)
	}
	dashboard, err := NewDashboard(application, nil, store, nil)
	if err != nil {
		t.Fatal(err)
	}
	seen := make(map[string]bool)
	for _, category := range Categories {
		if dashboard.Screen(category.ID) == nil {
			t.Errorf("category %q has no screen", category.ID)
		}
		seen[category.ID] = true
	}
	if len(seen) != 11 {
		t.Fatalf("expected 11 navigation screens, got %d", len(seen))
	}
}

func TestGlobalSearchFindsNodesThroughLinkedFindings(t *testing.T) {
	application := test.NewApp()
	store, err := repository.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err = (service.Service{Store: store}).LoadDemo(context.Background()); err != nil {
		t.Fatal(err)
	}
	dashboard, err := NewDashboard(application, nil, store, nil)
	if err != nil {
		t.Fatal(err)
	}
	dashboard.search = "stale credential reference"
	found := make(map[string]bool)
	for _, node := range dashboard.filteredNodes() {
		found[node.ID] = true
	}
	if !found["secret-store"] || !found["deploy-bot"] {
		t.Fatalf("global search did not match nodes through their linked finding: %#v", found)
	}
}

func TestAboutLinksStayWithinConfiguredRepository(t *testing.T) {
	allowed := []string{
		"https://github.com/NexzaDev/attackpath-visualizer",
		"https://github.com/NexzaDev/attackpath-visualizer/issues/new",
		"https://github.com/NexzaDev/attackpath-visualizer/blob/main/docs/USAGE.md",
		"https://github.com/NexzaDev/attackpath-visualizer/security/advisories/new",
	}
	for _, target := range allowed {
		if !AllowedRepositoryURL(target) {
			t.Errorf("expected configured link to pass: %s", target)
		}
	}
	blocked := []string{
		"http://github.com/NexzaDev/attackpath-visualizer",
		"https://github.com.evil.example/NexzaDev/attackpath-visualizer",
		"https://github.com/another/project",
		"https://github.com/NexzaDev/attackpath-visualizer.evil",
		"javascript:alert(1)",
	}
	for _, target := range blocked {
		if AllowedRepositoryURL(target) {
			t.Errorf("unexpectedly allowed external link: %s", target)
		}
	}
}

func TestSoftwareRendererProducesAllScreenImages(t *testing.T) {
	application := test.NewApp()
	output := t.TempDir()
	if err := RenderScreenshots(output, application); err != nil {
		t.Fatal(err)
	}
	for _, category := range Categories {
		if _, err := os.Stat(filepath.Join(output, category.ID+".png")); err != nil {
			t.Errorf("missing screenshot for %s: %v", category.ID, err)
		}
	}
}
