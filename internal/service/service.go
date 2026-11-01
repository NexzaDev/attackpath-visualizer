package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/NexzaDev/attackpath-visualizer/assets"
	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
	"github.com/NexzaDev/attackpath-visualizer/internal/ingestion"
	"github.com/NexzaDev/attackpath-visualizer/internal/notifications"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
)

type Service struct{ Store *repository.Store }

func (s Service) LoadDemo(ctx context.Context) error {
	data, err := assets.Files.ReadFile("demo.json")
	if err != nil {
		return fmt.Errorf("read embedded demo fixture: %w", err)
	}
	return s.Import(ctx, strings.NewReader(string(data)), "demo fixture")
}

func (s Service) Import(ctx context.Context, reader io.Reader, source string) error {
	if source == "" {
		source = "local JSON import"
	}
	graph, validation, err := ingestion.Read(reader)
	if err != nil {
		return s.validationFailure(ctx, source, err)
	}
	previous, err := s.Store.Graph(ctx)
	if err != nil {
		return fmt.Errorf("read existing graph: %w", err)
	}
	result := repository.ImportResult{SchemaVersion: validation.SchemaVersion, IdentityCount: validation.Identities, FindingCount: validation.Findings, MisconfigurationCount: validation.Misconfigurations, DocumentBytes: validation.DocumentBytes}
	if err = s.Store.ReplaceGraphWithResult(ctx, graph, source, result); err != nil {
		return err
	}
	return s.recordNewCriticalPaths(ctx, previous, graph)
}

func (s Service) validationFailure(ctx context.Context, source string, cause error) error {
	if err := s.Store.RecordImportFailure(ctx, source, cause.Error()); err != nil {
		return fmt.Errorf("%w; record validation result: %v", cause, err)
	}
	rules, err := s.Store.NotificationRules(ctx)
	if err != nil {
		return fmt.Errorf("%w; read notification rules: %v", cause, err)
	}
	rule, configured := rules["validation_failed"]
	if configured && rule.Enabled && 3 >= rule.MinimumSeverity {
		_, err = s.Store.AddNotification(ctx, notifications.Event{Type: "validation_failed", Severity: 3, Title: "Import failed validation", Detail: "The selected inventory could not be read or validated.", Related: "imports"}, 15*time.Minute)
		if err != nil {
			return fmt.Errorf("%w; save validation notification: %v", cause, err)
		}
	}
	return cause
}

func (s Service) recordNewCriticalPaths(ctx context.Context, previous, current domain.Graph) error {
	before, err := previous.AllPaths(16, 1000)
	if err != nil {
		return err
	}
	known := make(map[string]bool, len(before))
	for _, path := range before {
		if isCriticalPath(path) {
			known[pathSignature(path)] = true
		}
	}
	paths, err := current.AllPaths(16, 1000)
	if err != nil {
		return err
	}
	rules, err := s.Store.NotificationRules(ctx)
	if err != nil {
		return err
	}
	rule := rules["critical_path"]
	if !rule.Enabled {
		return nil
	}
	for _, path := range paths {
		if !isCriticalPath(path) {
			continue
		}
		signature := pathSignature(path)
		if known[signature] {
			continue
		}
		_, err = s.Store.AddNotification(ctx, notifications.Event{Type: "critical_path", Severity: 5, Title: "New critical potential path", Detail: "An import added a new high-impact relationship path.", Related: "paths", Key: signature}, 15*time.Minute)
		if err != nil {
			return fmt.Errorf("record new critical path: %w", err)
		}
		known[signature] = true
	}
	return nil
}

func isCriticalPath(path domain.Path) bool {
	return path.Impact >= 5 && path.Confidence >= 60 && path.Score >= 300
}

func pathSignature(path domain.Path) string {
	parts := make([]string, 0, len(path.Nodes)+len(path.Edges))
	for _, node := range path.Nodes {
		parts = append(parts, "node:"+node.ID)
	}
	for _, edge := range path.Edges {
		parts = append(parts, "edge:"+edge.ID)
	}
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return hex.EncodeToString(sum[:])
}
