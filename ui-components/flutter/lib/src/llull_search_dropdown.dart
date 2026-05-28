import 'package:flutter/material.dart';
import 'llull_search_controller.dart';
import 'types.dart';

class LlullSearchDropdown extends StatefulWidget {
  final String host;
  final String? authToken;
  final String index;
  final String placeholder;
  final int debounceMs;
  final int minChars;
  final int maxResults;
  final ValueChanged<LlullSearchResult>? onSelected;

  const LlullSearchDropdown({
    super.key,
    required this.host,
    this.authToken,
    this.index = '',
    this.placeholder = 'Search...',
    this.debounceMs = 200,
    this.minChars = 2,
    this.maxResults = 5,
    this.onSelected,
  });

  @override
  State<LlullSearchDropdown> createState() => _LlullSearchDropdownState();
}

class _LlullSearchDropdownState extends State<LlullSearchDropdown> {
  final controller = TextEditingController();
  late final _searchController = LlullSearchController(
    host: '',
    authToken: null,
    index: widget.index,
  );
  final _focusNode = FocusNode();
  bool _open = false;
  Timer? _timer;

  @override
  void initState() {
    super.initState();
    controller.addListener(_onChanged);
    _focusNode.addListener(() {
      if (!_focusNode.hasFocus) setState(() => _open = false);
    });
  }

  void _onChanged() {
    _timer?.cancel();
    _timer = Timer(Duration(milliseconds: widget.debounceMs), () async {
      if (controller.text.trim().length >= widget.minChars) {
        await _searchController.search(
          query: controller.text,
          hitsPerPage: widget.maxResults,
        );
        if (mounted) setState(() => _open = true);
      } else {
        setState(() => _open = false);
      }
    });
  }

  @override
  void dispose() {
    controller.dispose();
    _searchController.dispose();
    _focusNode.dispose();
    _timer?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Column(
      children: [
        TextField(
          controller: controller,
          focusNode: _focusNode,
          decoration: InputDecoration(
            hintText: widget.placeholder,
            border: OutlineInputBorder(borderRadius: BorderRadius.circular(24)),
            contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
          ),
        ),
        if (_open && _searchController.results.isNotEmpty)
          Card(
            margin: EdgeInsets.zero,
            elevation: 4,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: _searchController.results.map((r) {
                final title = (r.fields?['title'] as String?) ?? r.id;
                return ListTile(
                  dense: true,
                  title: Text(title, maxLines: 1, overflow: TextOverflow.ellipsis),
                  subtitle: Text('Score: ${r.score.toStringAsFixed(3)}'),
                  onTap: () {
                    widget.onSelected?.call(r);
                    controller.text = title;
                    setState(() => _open = false);
                  },
                );
              }).toList(),
            ),
          ),
      ],
    );
  }
}
