package engram

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/bubustack/bobrapet/pkg/storage"
	"github.com/bubustack/bubu-sdk-go/engram"
	"github.com/bubustack/json-filter-engram/pkg/config"
)

type JSONFilter struct {
	config config.Config
}

func New() *JSONFilter {
	return &JSONFilter{}
}

func (e *JSONFilter) Init(_ context.Context, cfg config.Config, _ *engram.Secrets) error {
	e.config = cfg
	return nil
}

func (e *JSONFilter) Process(
	ctx context.Context,
	execCtx *engram.ExecutionContext,
	inputs config.Inputs,
) (*engram.Result, error) {
	logger := execCtx.Logger()
	if inputs.Input == nil {
		return nil, fmt.Errorf("input is required")
	}
	if inputs.Select == nil {
		return nil, fmt.Errorf("select is required")
	}

	resolved := inputs.Input
	if requiresHydration(resolved) {
		if sm, err := storage.SharedManager(ctx); err != nil {
			return nil, fmt.Errorf("storage manager unavailable: %w", err)
		} else if sm != nil {
			resolved, err = sm.Hydrate(ctx, resolved)
			if err != nil {
				return nil, fmt.Errorf("failed to hydrate input: %w", err)
			}
		}
	}

	parseJSON := true
	if inputs.ParseJSONText != nil {
		parseJSON = *inputs.ParseJSONText
	}

	root := resolved
	if parseJSON {
		if parsed := normalizeInput(resolved); parsed != nil {
			root = parsed
		}
	}

	selected, err := selectFrom(root, inputs.Select)
	if err != nil {
		return nil, err
	}

	if logger != nil {
		logger.Info("json-filter selected value", slog.Any("keys", keysOf(selected)))
	}

	return engram.NewResultFrom(map[string]any{"resultSelected": selected}), nil
}

func requiresHydration(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		if hasHydrationRefKeys(typed) {
			return true
		}
		for _, nested := range typed {
			if requiresHydration(nested) {
				return true
			}
		}
	case []any:
		for _, nested := range typed {
			if requiresHydration(nested) {
				return true
			}
		}
	}
	return false
}

func hasHydrationRefKeys(values map[string]any) bool {
	for key := range values {
		switch key {
		case storage.StorageRefKey, "$bubuConfigMapRef", "$bubuSecretRef":
			return true
		}
	}
	return false
}

func normalizeInput(input any) any {
	switch typed := input.(type) {
	case map[string]any:
		if parsed := parseDataItemJSON(typed); parsed != nil {
			return parsed
		}
		if content, ok := typed["content"].([]any); ok {
			if parsed := parseContentItemsJSON(content); parsed != nil {
				return parsed
			}
		}
		if content, ok := typed["content"].(map[string]any); ok {
			if parsed := parseDataItemJSON(content); parsed != nil {
				return parsed
			}
		}
		if text, ok := typed["text"].(string); ok {
			if parsed, ok := parseJSONText(text); ok {
				return parsed
			}
		}
		return nil
	case []any:
		if parsed := parseContentItemsJSON(typed); parsed != nil {
			return parsed
		}
		return nil
	case string:
		if parsed, ok := parseJSONText(typed); ok {
			return parsed
		}
		return nil
	default:
		return nil
	}
}

func parseDataItemJSON(result map[string]any) any {
	raw, ok := result["data"]
	if !ok {
		return nil
	}
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	parsedItems := make([]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		text, ok := entry["text"].(string)
		if !ok {
			continue
		}
		parsed, ok := parseJSONText(text)
		if !ok {
			continue
		}
		parsedItems = append(parsedItems, parsed)
	}
	if len(parsedItems) == 0 {
		return nil
	}
	if len(parsedItems) == 1 {
		return parsedItems[0]
	}
	return parsedItems
}

func parseContentItemsJSON(items []any) any {
	if len(items) == 0 {
		return nil
	}
	parsedItems := make([]any, 0, len(items))
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		text, ok := entry["text"].(string)
		if !ok {
			continue
		}
		parsed, ok := parseJSONText(text)
		if !ok {
			continue
		}
		parsedItems = append(parsedItems, parsed)
	}
	if len(parsedItems) == 0 {
		return nil
	}
	if len(parsedItems) == 1 {
		return parsedItems[0]
	}
	return parsedItems
}

func parseJSONText(text string) (any, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return nil, false
	}
	if !strings.HasPrefix(trimmed, "{") && !strings.HasPrefix(trimmed, "[") {
		return nil, false
	}
	var parsed any
	if err := json.Unmarshal([]byte(trimmed), &parsed); err != nil {
		return nil, false
	}
	return parsed, true
}

func selectFrom(root any, sel *config.Selection) (map[string]any, error) {
	if sel == nil {
		return nil, fmt.Errorf("select is required")
	}
	items, _, err := resolveItems(root, sel)
	if err != nil {
		return nil, err
	}
	matchField := strings.TrimSpace(sel.MatchField)
	matched := findMatchedItem(items, matchField, sel.MatchValue)
	requireMatch := true
	if sel.RequireMatch != nil {
		requireMatch = *sel.RequireMatch
	}
	if matched == nil {
		if requireMatch {
			return nil, fmt.Errorf("no match for %s=%q", matchField, sel.MatchValue)
		}
		return map[string]any{}, nil
	}

	return buildSelectedValue(sel, matched)
}

func resolveItems(root any, sel *config.Selection) ([]any, string, error) {
	itemsSource := root
	itemsPath := strings.TrimSpace(sel.ItemsPath)
	if itemsPath != "" {
		value, ok := getPath(root, itemsPath)
		if !ok {
			return nil, "", fmt.Errorf("itemsPath not found: %s", itemsPath)
		}
		itemsSource = value
	}
	items, ok := itemsSource.([]any)
	if !ok {
		return nil, "", fmt.Errorf("itemsPath is not a list: %s", itemsPath)
	}
	return items, itemsPath, nil
}

func findMatchedItem(items []any, matchField, matchValue string) map[string]any {
	for _, item := range items {
		entry, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if matchField == "" {
			return entry
		}
		value, ok := getPath(entry, matchField)
		if ok && valuesEqual(value, matchValue) {
			return entry
		}
	}
	return nil
}

func buildSelectedValue(sel *config.Selection, matched map[string]any) (map[string]any, error) {
	selected := any(matched)
	if valuePath := strings.TrimSpace(sel.ValuePath); valuePath != "" {
		value, ok := getPath(matched, valuePath)
		if !ok {
			return nil, fmt.Errorf("valuePath not found: %s", sel.ValuePath)
		}
		selected = value
	}
	outputKey := strings.TrimSpace(sel.OutputKey)
	if outputKey == "" {
		outputKey = "value"
	}
	out := map[string]any{outputKey: selected}
	if sel.IncludeItem {
		out["item"] = matched
	}
	return out, nil
}

type pathToken struct {
	key   string
	index *int
}

func getPath(root any, path string) (any, bool) {
	if path == "" {
		return root, true
	}
	tokens, err := parsePath(path)
	if err != nil {
		return nil, false
	}
	current := root
	for _, tok := range tokens {
		if tok.key != "" {
			m, ok := current.(map[string]any)
			if !ok {
				return nil, false
			}
			val, ok := m[tok.key]
			if !ok {
				return nil, false
			}
			current = val
		}
		if tok.index != nil {
			arr, ok := current.([]any)
			if !ok {
				return nil, false
			}
			idx := *tok.index
			if idx < 0 || idx >= len(arr) {
				return nil, false
			}
			current = arr[idx]
		}
	}
	return current, true
}

func parsePath(path string) ([]pathToken, error) {
	segments := strings.Split(path, ".")
	tokens := make([]pathToken, 0, len(segments))
	for _, seg := range segments {
		if seg == "" {
			continue
		}
		for seg != "" {
			if seg[0] == '[' {
				end := strings.Index(seg, "]")
				if end == -1 {
					return nil, fmt.Errorf("invalid path: %s", path)
				}
				idx, err := strconv.Atoi(seg[1:end])
				if err != nil {
					return nil, fmt.Errorf("invalid index in path: %s", path)
				}
				tokens = append(tokens, pathToken{index: &idx})
				seg = seg[end+1:]
				continue
			}
			idx := strings.Index(seg, "[")
			if idx == -1 {
				tokens = append(tokens, pathToken{key: seg})
				seg = ""
				continue
			}
			if idx > 0 {
				tokens = append(tokens, pathToken{key: seg[:idx]})
			}
			seg = seg[idx:]
		}
	}
	return tokens, nil
}

func valuesEqual(val any, expected string) bool {
	switch typed := val.(type) {
	case string:
		return typed == expected
	case fmt.Stringer:
		return typed.String() == expected
	default:
		return fmt.Sprint(val) == expected
	}
}

func keysOf(m map[string]any) []string {
	if len(m) == 0 {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
