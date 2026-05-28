import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
import 'types.dart';

class LlullSearchController {
  final String host;
  final String? authToken;
  final http.Client _client = http.Client();

  List<LlullSearchResult> results = [];
  int totalHits = 0;
  bool loading = false;
  String? error;
  Completer? _abortCompleter;

  LlullSearchController({required this.host, this.authToken});

  Future<void> search({
    required String query,
    int page = 1,
    int hitsPerPage = 10,
    bool fuzzy = true,
    bool useWeight = false,
    double weightImpact = 0.3,
  }) async {
    if (query.trim().isEmpty) {
      results = [];
      totalHits = 0;
      return;
    }

    _abortCompleter?.complete();
    _abortCompleter = Completer();
    loading = true;
    error = null;

    try {
      final params = {
        'q': query,
        'page': page.toString(),
        'hits_per_page': hitsPerPage.toString(),
        'fuzzy': fuzzy ? 'true' : 'false',
      };
      if (useWeight) {
        params['use_weight'] = 'true';
        params['weight_impact'] = weightImpact.toString();
      }

      final uri = Uri.parse('$host/v1/search').replace(queryParameters: params);
      final headers = <String, String>{'Content-Type': 'application/json'};
      if (authToken != null) headers['Authorization'] = 'Bearer $authToken';

      final response = await _client.get(uri, headers: headers).timeout(
        const Duration(seconds: 10),
      );

      if (response.statusCode == 200) {
        final data = LlullSearchResponse.fromJson(
          jsonDecode(response.body) as Map<String, dynamic>,
        );
        results = data.hits;
        totalHits = data.totalHits;
      } else {
        error = 'HTTP ${response.statusCode}';
      }
    } catch (e) {
      error = e.toString();
    } finally {
      loading = false;
    }
  }

  void dispose() {
    _client.close();
    _abortCompleter?.complete();
  }
}
