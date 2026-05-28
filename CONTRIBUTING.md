# Contributing to Llull

We welcome contributions! Here's how to get started.

## Development Setup

```bash
git clone git@github.com:2mes4/llull-searchengine.git
cd llull-searchengine
go test ./... -v -race
```

## Project Structure

```
llull/
├── cmd/server/main.go              Entry point (IndexManager, BoltDB, multi-index)
├── internal/
│   ├── engine/                     Core search (IndexManager, trie, ranking, fuzzy, persist)
│   ├── api/                        HTTP handlers (chi router, multi-index routes)
│   ├── worker/                     Worker pool (buffered channel)
│   ├── datasource/                 Data source interface + connectors
│   └── seed/                       Seed data generator
├── data-sources/firestore/         Firebase Extension (push model)
├── data-sources/postgres/          PostgreSQL connector
├── data-sources/mysql/             MySQL connector
├── data-sources/mongodb/           MongoDB connector
├── skills/                         AI skills for contributors
├── ui-components/react/            React component library
├── ui-components/flutter/          Flutter widget library
├── web/                            Built-in Google-style search UI
└── deploy/                         Docker, Kubernetes, install docs
```

## Code Style

- **Go**: No comments unless requested. Use `fmt.Errorf` with `%w` for errors. `sync.RWMutex` for trie concurrency
- **React**: Functional components with TypeScript. Hooks, no class components. Export types. Accept `index` prop for multi-index
- **Flutter**: Widgets with `const` constructors. `StatelessWidget` > `StatefulWidget` when possible. Accept `index` parameter
- **All**: English language, no emojis unless requested

## Testing

```bash
# Go engine tests (always with -race)
go test ./... -v -race

# React
cd ui-components/react && npm test

# Flutter
cd ui-components/flutter && flutter test
```

## Data Source Contribution

1. Implement `internal/datasource/datasource.go` interface
2. Add connector in `internal/datasource/<name>/`
3. Include `Index` field in `Config` to specify target Llull index
4. Document in `data-sources/<name>/README.md`
5. Include README with configuration table
6. Add tests

See `skills/llull-searchengine-datasources-creator/SKILL.md` for the full guide.

## Frontend Contribution

1. React components go in `ui-components/react/src/`
2. Flutter widgets go in `ui-components/flutter/lib/`
3. Export from `src/index.tsx` or `lib/main.dart`
4. Include TypeScript/TypeDefs for React
5. Accept optional `index` prop/parameter for multi-index support
6. Follow existing patterns

## Pull Request Process

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/my-feature`)
3. Commit changes (`git commit -m 'feat: add my feature'`)
4. Push branch (`git push origin feat/my-feature`)
5. Open a Pull Request

## License

By contributing, you agree that your contributions will be licensed under the project's existing license.
