import React, { useState, useCallback, useEffect } from 'react';
import { useLlullSearch } from '../hooks/useLlullSearch';
import type { LlullSearchResult } from '../types';

interface LlullSearchResultsProps {
  host: string;
  index?: string;
  authToken?: string;
  renderCard?: (result: LlullSearchResult, query: string) => React.ReactNode;
  hitsPerPage?: number;
}

export function LlullSearchResults({
  host,
  index,
  authToken,
  renderCard,
  hitsPerPage = 10,
}: LlullSearchResultsProps) {
  const [query, setQuery] = useState('');
  const [page, setPage] = useState(1);
  const { search, results, totalHits, loading, error } = useLlullSearch({ host, index, authToken });

  const handleInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setQuery(e.target.value);
    setPage(1);
  }, []);

  useEffect(() => {
    if (query.trim()) {
      search(query, { page, hitsPerPage });
    }
  }, [query, page, hitsPerPage, search]);

  const totalPages = Math.ceil(totalHits / hitsPerPage);

  return (
    <div>
      <input
        value={query}
        onChange={handleInput}
        placeholder="Search..."
        style={{
          width: '100%', padding: '12px 16px', fontSize: 16,
          border: '1px solid #dfe1e5', borderRadius: 24, outline: 'none',
          marginBottom: 16, boxSizing: 'border-box',
        }}
      />

      {error && <div style={{ color: 'red', padding: 12 }}>{error}</div>}
      {loading && <div style={{ padding: 12, color: '#666' }}>Searching...</div>}

      {!loading && !error && results.map((r) => (
        renderCard ? (
          renderCard(r, query)
        ) : (
          <DefaultCard key={r.id} result={r} query={query} />
        )
      ))}

      {totalPages > 1 && (
        <div style={{ display: 'flex', justifyContent: 'center', gap: 8, marginTop: 16 }}>
          <button disabled={page <= 1} onClick={() => setPage(p => p - 1)}
            style={btnStyle}>Prev</button>
          <span style={{ padding: '6px 12px', color: '#666', fontSize: 14 }}>
            {page} / {totalPages}
          </span>
          <button disabled={page >= totalPages} onClick={() => setPage(p => p + 1)}
            style={btnStyle}>Next</button>
        </div>
      )}
    </div>
  );
}

function DefaultCard({ result, query }: { result: LlullSearchResult; query: string }) {
  const fields = result.fields || {};
  const title: string = fields.title || result.id;
  const content: string = (fields.content || '').slice(0, 280);

  return (
    <div style={{
      padding: '14px 16px', marginBottom: 8,
      border: '1px solid #e8eaed', borderRadius: 8,
    }}>
      <h3 style={{ margin: '0 0 4px', fontSize: 16, fontWeight: 500 }}>
        {highlight(title, query)}
      </h3>
      <p style={{ margin: 0, color: '#4d5156', fontSize: 14, lineHeight: 1.5 }}>
        {highlight(content, query)}
      </p>
      <div style={{ marginTop: 6, fontSize: 12, color: '#70757a', display: 'flex', gap: 12 }}>
        <span>Weight: {(result.weight * 100).toFixed(0)}%</span>
        <span>Score: {result.score.toFixed(3)}</span>
      </div>
    </div>
  );
}

function highlight(text: string, query: string): React.ReactNode {
  if (!text || !query) return text;
  const lowerText = text.toLowerCase();
  const lowerQuery = query.toLowerCase();
  const parts: React.ReactNode[] = [];
  let lastIdx = 0;
  let idx = lowerText.indexOf(lowerQuery, lastIdx);
  while (idx !== -1) {
    parts.push(text.slice(lastIdx, idx));
    parts.push(<strong key={idx}>{text.slice(idx, idx + query.length)}</strong>);
    lastIdx = idx + query.length;
    idx = lowerText.indexOf(lowerQuery, lastIdx);
  }
  parts.push(text.slice(lastIdx));
  return parts;
}

const btnStyle: React.CSSProperties = {
  padding: '6px 14px', border: '1px solid #dadce0',
  borderRadius: 6, background: '#fff', cursor: 'pointer', fontSize: 13,
};
