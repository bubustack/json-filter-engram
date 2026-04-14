# 🧮 JSON Filter Engram

Selects structured data from JSON payloads (inline or $bubuStorageRef-offloaded)
and returns a compact result object configurable by selectors.

## 🌟 Highlights

- Works with inline or offloaded JSON using the same schema.
- Supports `matchField`, `valuePath`, `outputKey`, and conditional inclusion.
- Returns deterministic inline metadata that downstream steps consume directly.

## 🚀 Quick Start

```bash
go test ./...
```

Apply `Engram.yaml`, then reference the template in your Story step and supply
the selector inputs shown below.

## ⚙️ Configuration (`Engram.spec.with`)

This engram currently exposes no component-level `Engram.spec.with` options.
Its `configSchema` is empty, so behavior is controlled entirely by per-execution
inputs.

## 📥 Inputs

- `input`: JSON data (inline or storage ref).
- `select.itemsPath`: JSONPath to the array of items.
- `select.matchField`: field within each item to compare.
- `select.matchValue`: value that must match for inclusion.
- `select.valuePath`: subfield whose value is captured.
- `select.outputKey`: key name for the extracted result.
- `select.includeItem`: includes the matched item when `true`.
- `select.requireMatch`: fails the step when `true` and no match is found.
- `parseJSONText`: parses stringified JSON payloads when `true`.

Example:

```yaml
select:
  itemsPath: channels
  matchField: name
  matchValue: daily-digest
  outputKey: channelId
```

## 📤 Outputs

Returns an object under `resultSelected` with the configured `outputKey` and,
when `includeItem` is `true`, the matching item under `item`.

## 🧪 Local Development

- `go test ./...` – Run unit tests covering inline/offload flows.
- `go vet ./...` – Ensure schema compliance before publishing.

## 🤝 Community & Support

- [Contributing](./CONTRIBUTING.md)
- [Support](./SUPPORT.md)
- [Security Policy](./SECURITY.md)
- [Code of Conduct](./CODE_OF_CONDUCT.md)
- [Discord](https://discord.gg/dysrB7D8H6)

## 📄 License

Copyright 2025 BubuStack.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
