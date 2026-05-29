# Llull UI Components

Plug-and-play search UI components for React and Flutter. Connect to any Llull server instance with a single configuration line.

## React (`llull-search-components`)

```bash
npm install llull-search-components
```

### `<LlullSearchDropdown />`

Search-as-you-type autocomplete dropdown. Shows matching results as the user types.

```jsx
import { LlullSearchDropdown } from 'llull-search-components';

function MyComponent() {
  return <LlullSearchDropdown
    host="http://localhost:8180"
    placeholder="Search documents..."
    debounceMs={200}
    minChars={2}
    maxResults={5}
    onSelected={(result) => console.log('Selected:', result.id)}
  />;
}
```

### `useLlullSearch()` (Headless hook)

Returns an array of search results with full control over rendering.

```jsx
import { useLlullSearch } from 'llull-search-components';

function MyComponent() {
  const { search, results, totalHits, loading, error } = useLlullSearch({
    host: 'http://localhost:8180'
  });

  return <>
    <input onChange={e => search(e.target.value)} />
    {results.map(r => <div key={r.id}>{r.id} — Score: {r.score.toFixed(3)}</div>)}
  </>;
}
```

### `<LlullSearchResults />`

Full search interface with input, paginated results, and score display.

```jsx
import { LlullSearchResults } from 'llull-search-components';

function MyPage() {
  return <LlullSearchResults host="http://localhost:8180" />;
}
```

Custom card rendering:

```jsx
<LlullSearchResults
  host="http://localhost:8180"
  renderCard={(result, query) => (
    <MyCustomCard result={result} />
  )}
/>
```

## Flutter (`llull_search_components`)

```yaml
dependencies:
  llull_search_components: ^0.1.0
```

### `LlullSearchDropdown`

```dart
LlullSearchDropdown(
  host: 'http://localhost:8180',
  onSelected: (result) => print('Selected: ${result.id}'),
)
```

### `LlullSearchController` (Headless)

```dart
final controller = LlullSearchController(host: 'http://localhost:8180');
await controller.search(query: 'ciencia');
print(controller.results);
```

### `LlullSearchResults`

```dart
LlullSearchResults(
  host: 'http://localhost:8180',
  cardBuilder: (result, query) => MyCustomCard(result: result),
)
```

## Design Decisions

- **Zero framework lock-in**: React and Flutter are separate packages with no shared dependencies
- **Minimal API surface**: Components accept only a `host` URL. Configurable via optional props
- **Abort handling**: All requests abort when the query changes (prevents race conditions)
- **Match highlighting**: Search terms are highlighted in bold in titles and content snippets
- **Score transparency**: Every result displays its relevance score and platform weight
