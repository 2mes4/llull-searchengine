import { useState, useCallback, useRef, useEffect } from 'react';
import type { LlullSearchResponse, LlullSearchResult } from '../types';

interface UseLlullSearchConfig {
  host: string;
  index?: string;
  authToken?: string;
}

export function useLlullSearch(config: UseLlullSearchConfig) {
  const [results, setResults] = useState<LlullSearchResult[]>([]);
  const [totalHits, setTotalHits] = useState(0);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const search = useCallback(async (query: string, options?: {
    page?: number; hitsPerPage?: number; fuzzy?: boolean;
    useWeight?: boolean; weightImpact?: number;
  }) => {
    if (!query.trim()) {
      setResults([]); setTotalHits(0); return;
    }

    if (abortRef.current) abortRef.current.abort();
    abortRef.current = new AbortController();

    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams({
        q: query,
        page: String(options?.page || 1),
        hits_per_page: String(options?.hitsPerPage || 10),
        fuzzy: options?.fuzzy !== false ? 'true' : 'false',
      });
      if (options?.useWeight) {
        params.set('use_weight', 'true');
        params.set('weight_impact', String(options.weightImpact || 0.3));
      }

      const basePath = config.index ? `/v1/${config.index}/search` : '/v1/search';
      const res = await fetch(`${config.host}${basePath}?${params}`, {
        signal: abortRef.current.signal,
        headers: config.authToken
          ? { Authorization: `Bearer ${config.authToken}` }
          : undefined,
      });

      if (!res.ok) throw new Error(`HTTP ${res.status}`);

      const data: LlullSearchResponse = await res.json();
      setResults(data.hits);
      setTotalHits(data.total_hits);
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        setError(err.message);
      }
    } finally {
      setLoading(false);
    }
  }, [config.host, config.authToken]);

  const clear = useCallback(() => {
    setResults([]);
    setTotalHits(0);
    setError(null);
  }, []);

  useEffect(() => {
    return () => { if (abortRef.current) abortRef.current.abort(); };
  }, []);

  return { search, results, totalHits, loading, error, clear };
}
