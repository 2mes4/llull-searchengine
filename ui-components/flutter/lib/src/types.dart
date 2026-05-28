class LlullSearchResult {
  final String id;
  final double score;
  final double weight;
  final Map<String, dynamic>? fields;

  LlullSearchResult({
    required this.id,
    required this.score,
    required this.weight,
    this.fields,
  });

  factory LlullSearchResult.fromJson(Map<String, dynamic> json) {
    return LlullSearchResult(
      id: json['id'] as String,
      score: (json['score'] as num).toDouble(),
      weight: (json['weight'] as num?)?.toDouble() ?? 0.0,
      fields: json['fields'] as Map<String, dynamic>?,
    );
  }
}

class LlullSearchResponse {
  final List<LlullSearchResult> hits;
  final int totalHits;
  final int page;
  final int nbPages;
  final int hitsPerPage;
  final String query;

  LlullSearchResponse({
    required this.hits,
    required this.totalHits,
    required this.page,
    required this.nbPages,
    required this.hitsPerPage,
    required this.query,
  });

  factory LlullSearchResponse.fromJson(Map<String, dynamic> json) {
    return LlullSearchResponse(
      hits: (json['hits'] as List)
          .map((e) => LlullSearchResult.fromJson(e as Map<String, dynamic>))
          .toList(),
      totalHits: json['total_hits'] as int,
      page: json['page'] as int,
      nbPages: json['nb_pages'] as int,
      hitsPerPage: json['hits_per_page'] as int,
      query: json['query'] as String,
    );
  }
}
