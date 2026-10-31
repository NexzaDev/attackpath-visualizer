package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/NexzaDev/attackpath-visualizer/internal/domain"
	"github.com/NexzaDev/attackpath-visualizer/internal/repository"
)

const maxImportBytes = 8 << 20

type Service struct{ Store *repository.Store }

func (s Service) LoadDemo(ctx context.Context, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("open demo fixture: %w", err)
	}
	defer file.Close()
	return s.Import(ctx, file, "demo fixture")
}

func (s Service) Import(ctx context.Context, reader io.Reader, source string) error {
	data, err := io.ReadAll(io.LimitReader(reader, maxImportBytes+1))
	if err != nil {
		return fmt.Errorf("read import: %w", err)
	}
	if len(data) > maxImportBytes {
		return fmt.Errorf("import exceeds the %d byte limit", maxImportBytes)
	}
	var graph domain.Graph
	decoderErr := json.Unmarshal(data, &graph)
	if decoderErr != nil {
		return fmt.Errorf("parse graph JSON: %w", decoderErr)
	}
	if err := graph.Validate(); err != nil {
		return fmt.Errorf("validate graph: %w", err)
	}
	return s.Store.ReplaceGraph(ctx, graph, source)
}
