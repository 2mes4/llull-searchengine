import 'package:flutter/material.dart';
import 'llull_search_controller.dart';
import 'types.dart';

class LlullSearchResults extends StatefulWidget {
  final String host;
  final String? authToken;
  final int hitsPerPage;
  final Widget Function(LlullSearchResult result, String query)? cardBuilder;

  const LlullSearchResults({
    super.key,
    required this.host,
    this.authToken,
    this.hitsPerPage = 10,
    this.cardBuilder,
  });

  @override
  State<LlullSearchResults> createState() => _LlullSearchResultsState();
}

class _LlullSearchResultsState extends State<LlullSearchResults> {
  final _queryController = TextEditingController();
  final _searchController = LlullSearchController(host: '');
  int _page = 1;

  @override
  void initState() {
    super.initState();
    _queryController.addListener(() {
      _page = 1;
      _doSearch();
    });
  }

  @override
  void dispose() {
    _queryController.dispose();
    _searchController.dispose();
    super.dispose();
  }

  void _doSearch() {
    final q = _queryController.text.trim();
    if (q.isEmpty) return;
    _searchController.search(
      query: q,
      page: _page,
      hitsPerPage: widget.hitsPerPage,
    );
    setState(() {});
  }

  int get _totalPages =>
      _searchController.totalHits ~/ widget.hitsPerPage +
      (_searchController.totalHits % widget.hitsPerPage > 0 ? 1 : 0);

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        TextField(
          controller: _queryController,
          decoration: InputDecoration(
            hintText: 'Search...',
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(24)),
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          ),
        ),
        const SizedBox(height: 12),
        if (_searchController.error != null)
          Padding(
            padding: const EdgeInsets.all(12),
            child: Text('Error: ${_searchController.error}', style: const TextStyle(color: Colors.red)),
          ),
        if (_searchController.loading)
          const Padding(
            padding: EdgeInsets.all(12),
            child: CircularProgressIndicator(),
          ),
        if (!_searchController.loading && _searchController.error == null)
          ..._searchController.results.map((r) {
            final q = _queryController.text.trim();
            return widget.cardBuilder?.call(r, q) ?? _defaultCard(r, q);
          }),
        if (_totalPages > 1)
          Padding(
            padding: const EdgeInsets.symmetric(vertical: 16),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                IconButton(
                  icon: const Icon(Icons.chevron_left),
                  onPressed: _page <= 1 ? null : () {
                    _page--;
                    _doSearch();
                  },
                ),
                Text('$_page / $_totalPages'),
                IconButton(
                  icon: const Icon(Icons.chevron_right),
                  onPressed: _page >= _totalPages ? null : () {
                    _page++;
                    _doSearch();
                  },
                ),
              ],
            ),
          ),
      ],
    );
  }

  Widget _defaultCard(LlullSearchResult r, String query) {
    final fields = r.fields ?? {};
    final title = (fields['title'] as String?) ?? r.id;
    final content = ((fields['content'] as String?) ?? '').take(200);

    return Card(
      margin: const EdgeInsets.only(bottom: 8),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(title, style: const TextStyle(fontSize: 16, fontWeight: FontWeight.w500)),
            const SizedBox(height: 4),
            Text(content.toString(), style: const TextStyle(fontSize: 14, color: Colors.grey)),
            const SizedBox(height: 6),
            Row(
              children: [
                Text('Weight: ${(r.weight * 100).toInt()}%', style: const TextStyle(fontSize: 12, color: Colors.grey)),
                const SizedBox(width: 12),
                Text('Score: ${r.score.toStringAsFixed(3)}', style: const TextStyle(fontSize: 12, color: Colors.grey)),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
