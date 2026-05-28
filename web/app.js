(function () {
  var API_BASE = '/v1';
  var DEBOUNCE_MS = 300;

  var searchInput = document.getElementById('search-input');
  var resultsEl = document.getElementById('results');
  var statsEl = document.getElementById('result-stats');
  var paginationEl = document.getElementById('pagination');
  var emptyEl = document.getElementById('empty-state');
  var heroEl = document.getElementById('hero');
  var useWeightEl = document.getElementById('use-weight');
  var weightGroup = document.getElementById('weight-group');
  var weightImpactEl = document.getElementById('weight-impact');
  var weightValueEl = document.getElementById('weight-value');
  var fuzzyEl = document.getElementById('fuzzy-search');
  var indexSelect = document.getElementById('index-select');

  var currentQuery = '';
  var currentPage = 1;
  var debounceTimer = null;
  var abortController = null;
  var currentHits = [];

  function openModal(hit) {
    var f = hit.fields || {};
    var title = f.title || hit.id;
    var content = f.content || '';
    var source = f.source || '';
    var weight = hit.weight != null ? hit.weight : (f.weight || 0);
    var weightPct = Math.round(weight * 100);

    document.getElementById('modal-title').innerHTML = highlight(title, currentQuery);
    document.getElementById('modal-source').textContent = source;
    document.getElementById('modal-id').textContent = hit.id;
    document.getElementById('modal-weight').textContent = weightPct + '%';
    document.getElementById('modal-score').textContent = Math.round(hit.score * 1000) / 1000;
    document.getElementById('modal-content').innerHTML = highlight(content, currentQuery);
    document.getElementById('modal-overlay').classList.add('open');
  }

  function closeModal() {
    document.getElementById('modal-overlay').classList.remove('open');
  }

  function init() {
    searchInput.addEventListener('input', function () {
      clearTimeout(debounceTimer);
      debounceTimer = setTimeout(function () { doSearch(1); }, DEBOUNCE_MS);
    });
    searchInput.addEventListener('keydown', function (e) {
      if (e.key === 'Enter') { clearTimeout(debounceTimer); doSearch(1); }
    });
    useWeightEl.addEventListener('change', function () {
      weightGroup.style.opacity = useWeightEl.checked ? '1' : '0.4';
      weightGroup.style.pointerEvents = useWeightEl.checked ? 'auto' : 'none';
      doSearch(1);
    });
    weightImpactEl.addEventListener('input', function () {
      weightValueEl.textContent = weightImpactEl.value + '%';
    });
    weightImpactEl.addEventListener('change', function () { doSearch(1); });
    fuzzyEl.addEventListener('change', function () { doSearch(1); });
    indexSelect.addEventListener('change', function () { doSearch(1); });

    loadStats();
    loadIndices();
    setInterval(loadStats, 3000);
  }

  function loadIndices() {
    fetch(API_BASE + '/indices')
      .then(function (r) { return r.json(); })
      .then(function (d) {
        if (!d.indices) return;
        var sel = indexSelect;
        sel.innerHTML = '';
        var names = Object.keys(d.indices);
        names.forEach(function (name) {
          var opt = document.createElement('option');
          opt.value = name;
          opt.textContent = name + ' (' + d.indices[name].docs + ')';
          sel.appendChild(opt);
        });
      })
      .catch(function () {});
  }

  function loadStats() {
    fetch(API_BASE + '/health')
      .then(function (r) { return r.json(); })
      .then(function (d) {
        document.getElementById('stat-docs').textContent = d.docs_indexed || '0';
        document.getElementById('stat-source').textContent = d.data_source || '—';
        document.getElementById('stat-memory').textContent = (d.total_memory_mb || d.memory_mb || '?') + ' MB';
        document.getElementById('stat-queue').textContent = d.queue_length || '0';
        document.getElementById('stat-index').textContent = indexSelect.value || d.default_index || 'default';
      })
      .catch(function () {});
  }

  function doSearch(page) {
    var q = searchInput.value.trim();
    currentPage = page || 1;
    if (!q) {
      resultsEl.innerHTML = '';
      statsEl.textContent = '';
      paginationEl.innerHTML = '';
      emptyEl.style.display = 'none';
      heroEl.style.display = '';
      return;
    }
    emptyEl.style.display = 'none';
    heroEl.style.display = 'none';
    if (abortController) abortController.abort();
    abortController = new AbortController();
    resultsEl.innerHTML = '<div class="loading">Searching...</div>';
    statsEl.textContent = '';
    currentQuery = q;

    var params = new URLSearchParams({
      q: q, page: currentPage, hits_per_page: 10,
      fuzzy: fuzzyEl.checked ? 'true' : 'false'
    });
    if (useWeightEl.checked) {
      params.set('use_weight', 'true');
      params.set('weight_impact', (parseInt(weightImpactEl.value) / 100).toString());
    }

    var idx = indexSelect.value;
    var searchSuffix = idx && idx !== 'default' ? '/' + idx + '/search' : '/search';
    fetch(API_BASE + searchSuffix + '?' + params.toString(), { signal: abortController.signal })
      .then(function (r) { return r.json(); })
      .then(function (data) { render(data); })
      .catch(function (err) {
        if (err.name !== 'AbortError') {
          resultsEl.innerHTML = '<div class="empty-state"><h3>Connection error</h3></div>';
        }
      });
  }

  function highlight(text, query) {
    if (!text || !query) return esc(text || '');
    var tokens = query.toLowerCase().split(/\s+/).filter(Boolean);
    var lower = text.toLowerCase();
    var parts = [];
    var i = 0;
    while (i < text.length) {
      var found = false;
      for (var t = 0; t < tokens.length; t++) {
        var tok = tokens[t];
        if (lower.substr(i, tok.length) === tok) {
          parts.push('<strong>' + esc(text.substr(i, tok.length)) + '</strong>');
          i += tok.length;
          found = true;
          break;
        }
      }
      if (!found) {
        parts.push(esc(text[i]));
        i++;
      }
    }
    return parts.join('');
  }

  function esc(s) {
    var d = document.createElement('div');
    d.appendChild(document.createTextNode(s));
    return d.innerHTML;
  }

  function truncate(text, max) {
    if (!text) return '';
    if (text.length <= max) return text;
    return text.substr(0, max) + '...';
  }

  function render(data) {
    if (!data.hits || data.hits.length === 0) {
      resultsEl.innerHTML = '';
      paginationEl.innerHTML = '';
      statsEl.textContent = '';
      emptyEl.style.display = '';
      return;
    }

    statsEl.textContent = 'About ' + data.total_hits + ' results' +
      ' (' + data.query_time + ' \u00b5s)' +
      (data.page > 1 ? ' - Page ' + data.page : '');

    var html = '';
    currentHits = data.hits;
    data.hits.forEach(function (hit, idx) {
      var f = hit.fields || {};
      var title = f.title || hit.id;
      var content = f.content || '';
      var source = f.source || '';
      var weight = hit.weight != null ? hit.weight : (f.weight || 0);
      var score = hit.score || 0;

      var snippet = truncate(content, 280);
      var highlighted = highlight(snippet, currentQuery);
      var highlightedTitle = highlight(truncate(title, 80), currentQuery);

      var weightPct = Math.round(weight * 100);
      var scoreStr = Math.round(score * 1000) / 1000;

      html += '<div class="result" onclick="window.__openModal(' + idx + ')" style="cursor:pointer;">';
      html += '<div class="url-line">';
      html += '<span class="source">' + esc(source) + '</span>';
      html += '<span class="doc-id">' + esc(hit.id) + '</span>';
      html += '</div>';
      html += '<h3><a onclick="window.__openModal(' + idx + ');return false">' + highlightedTitle + '</a></h3>';
      html += '<div class="snippet">' + highlighted + '</div>';
      html += '<div class="meta-row">';
      html += '<span>weight: ' + weightPct + '% <span class="weight-bar"><span class="weight-fill" style="width:' + weightPct + '%"></span></span></span>';
      html += '<span>score: ' + scoreStr + '</span>';
      html += '</div>';
      html += '</div>';
    });

    resultsEl.innerHTML = html;
    renderPagination(data);
  }

  function renderPagination(data) {
    if (data.nb_pages <= 1) { paginationEl.innerHTML = ''; return; }

    var html = '<div class="google-nav">';
    html += '<button class="page-btn" ' + (data.page <= 1 ? 'disabled' : '') + ' onclick="window.__prev()">&#8249;</button>';

    var start = Math.max(1, data.page - 5);
    var end = Math.min(data.nb_pages, data.page + 5);
    for (var i = start; i <= end; i++) {
      html += '<span class="page-num' + (i === data.page ? ' active' : '') + '">' + i + '</span>';
    }

    html += '<button class="page-btn" ' + (data.page >= data.nb_pages ? 'disabled' : '') + ' onclick="window.__next()">&#8250;</button>';
    html += '</div>';

    paginationEl.innerHTML = html;
  }

  window.__prev = function () { if (currentPage > 1) doSearch(currentPage - 1); };
  window.__next = function () { doSearch(currentPage + 1); };
  window.__triggerSearch = function () { doSearch(1); };
  window.__openModal = function (idx) { if (currentHits[idx]) openModal(currentHits[idx]); };
  window.closeModal = closeModal;
  window.__copy = function (btn) {
    var text = btn.parentElement.textContent.replace(/^Copia/, '').trim();
    navigator.clipboard.writeText(text).then(function () {
      btn.textContent = 'Copiat!';
      setTimeout(function () { btn.textContent = 'Copia'; }, 1500);
    });
  };

  init();
})();
