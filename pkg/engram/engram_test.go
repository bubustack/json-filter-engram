package engram

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/bubustack/bobrapet/pkg/storage"
	sdkengram "github.com/bubustack/bubu-sdk-go/engram"
	"github.com/bubustack/core/contracts"
	"github.com/bubustack/json-filter-engram/pkg/config"
)

func TestSelectFromMCPContent(t *testing.T) {
	input := map[string]any{
		"contentType": "json",
		"data": []any{
			map[string]any{
				"type": "text",
				"text": `{"ok":true,"channels":[{"id":"C123","name":"daily-digest"}]}`,
			},
		},
	}

	parsed := normalizeInput(input)
	if parsed == nil {
		t.Fatalf("expected parsed payload")
	}

	sel := &config.Selection{
		ItemsPath:  "channels",
		MatchField: "name",
		MatchValue: "daily-digest",
		ValuePath:  "id",
		OutputKey:  "channelId",
	}

	selected, err := selectFrom(parsed, sel)
	if err != nil {
		t.Fatalf("selectFrom error: %v", err)
	}
	if selected["channelId"] != "C123" {
		t.Fatalf("expected channelId C123, got %#v", selected)
	}
}

func TestSelectFromListRoot(t *testing.T) {
	root := []any{
		map[string]any{"id": "1", "name": "alpha"},
		map[string]any{"id": "2", "name": "beta"},
	}

	sel := &config.Selection{
		MatchField: "name",
		MatchValue: "beta",
		ValuePath:  "id",
		OutputKey:  "id",
	}

	selected, err := selectFrom(root, sel)
	if err != nil {
		t.Fatalf("selectFrom error: %v", err)
	}
	if selected["id"] != "2" {
		t.Fatalf("expected id 2, got %#v", selected)
	}
}

func TestProcessInlineInputSkipsStorageManagerInitialization(t *testing.T) {
	t.Setenv(contracts.StorageProviderEnv, "invalid-provider")
	storage.ResetSharedManagerCacheForTests()
	defer storage.ResetSharedManagerCacheForTests()

	filter := New()
	if err := filter.Init(context.Background(), config.Config{}, nil); err != nil {
		t.Fatalf("Init returned error: %v", err)
	}

	execCtx := sdkengram.NewExecutionContext(
		slog.New(slog.NewTextHandler(io.Discard, nil)),
		nil,
		sdkengram.StoryInfo{},
	)

	result, err := filter.Process(context.Background(), execCtx, config.Inputs{
		Input: []any{
			map[string]any{"id": "1", "name": "alpha"},
			map[string]any{"id": "2", "name": "beta"},
		},
		Select: &config.Selection{
			MatchField: "name",
			MatchValue: "beta",
			ValuePath:  "id",
			OutputKey:  "id",
		},
	})
	if err != nil {
		t.Fatalf("Process returned error: %v", err)
	}

	data, ok := result.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected map output, got %T", result.Data)
	}
	selected, ok := data["resultSelected"].(map[string]any)
	if !ok {
		t.Fatalf("expected resultSelected object, got %#v", data["resultSelected"])
	}
	if selected["id"] != "2" {
		t.Fatalf("expected id 2, got %#v", selected["id"])
	}
}
