export interface LlullSearchResult {
  id: string;
  score: number;
  weight: number;
  fields?: Record<string, any>;
}

export interface LlullSearchResponse {
  hits: LlullSearchResult[];
  total_hits: number;
  page: number;
  nb_pages: number;
  hits_per_page: number;
  query: string;
}

export interface LlullSearchOptions {
  host: string;
  query: string;
  page?: number;
  hitsPerPage?: number;
  fuzzy?: boolean;
  useWeight?: boolean;
  weightImpact?: number;
}
