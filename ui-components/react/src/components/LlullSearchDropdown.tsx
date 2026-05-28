import React, { useState, useEffect, useRef, useCallback } from 'react';
import { useLlullSearch } from '../hooks/useLlullSearch';

interface LlullSearchDropdownProps {
  host: string;
  index?: string;
  placeholder?: string;
  debounceMs?: number;
  minChars?: number;
  maxResults?: number;
  onSelected?: (result: any) => void;
  authToken?: string;
}

export function LlullSearchDropdown({
  host,
  index,
  placeholder = 'Search...',
  debounceMs = 200,
  minChars = 2,
  maxResults = 5,
  onSelected,
  authToken,
}: LlullSearchDropdownProps) {
  const [query, setQuery] = useState('');
  const [open, setOpen] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout>>();
  const { search, results, loading } = useLlullSearch({ host, index, authToken });

  const handleInput = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value;
    setQuery(val);
    if (timerRef.current) clearTimeout(timerRef.current);
    timerRef.current = setTimeout(() => {
      if (val.trim().length >= minChars) {
        search(val, { hitsPerPage: maxResults });
        setOpen(true);
      } else {
        setOpen(false);
      }
    }, debounceMs);
  }, [minChars, debounceMs, search, maxResults]);

  const handleSelect = useCallback((result: any) => {
    setQuery(result.fields?.title || result.id);
    setOpen(false);
    onSelected?.(result);
  }, [onSelected]);

  return (
    <div style={{ position: 'relative', width: '100%', maxWidth: 400 }}>
      <input
        value={query}
        onChange={handleInput}
        onFocus={() => results.length > 0 && setOpen(true)}
        onBlur={() => setTimeout(() => setOpen(false), 200)}
        placeholder={placeholder}
        style={{
          width: '100%',
          padding: '10px 14px',
          fontSize: 14,
          border: '1px solid #ddd',
          borderRadius: 8,
          outline: 'none',
          boxSizing: 'border-box',
        }}
      />
      {open && (loading || results.length > 0) && (
        <div style={{
          position: 'absolute', top: '100%', left: 0, right: 0,
          background: '#fff', border: '1px solid #ddd', borderRadius: 8,
          marginTop: 4, boxShadow: '0 4px 12px rgba(0,0,0,0.1)',
          zIndex: 1000, maxHeight: 300, overflowY: 'auto',
        }}>
          {results.map((r) => (
            <div key={r.id}
              onMouseDown={() => handleSelect(r)}
              style={{
                padding: '10px 14px', cursor: 'pointer',
                borderBottom: '1px solid #f0f0f0', fontSize: 13,
              }}
            >
              <div style={{ fontWeight: 500 }}>
                {highlight(r.fields?.title || r.id, query)}
              </div>
              <div style={{ color: '#666', fontSize: 12, marginTop: 2 }}>
                Score: {r.score.toFixed(3)} &middot; Weight: {(r.weight * 100).toFixed(0)}%
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function highlight(text: string, query: string): React.ReactNode {
  if (!text || !query) return text;
  const lowerText = text.toLowerCase();
  const lowerQuery = query.toLowerCase();
  const idx = lowerText.indexOf(lowerQuery);
  if (idx === -1) return text;
  return (
    <>
      {text.slice(0, idx)}
      <strong>{text.slice(idx, idx + query.length)}</strong>
      {text.slice(idx + query.length)}
    </>
  );
}
