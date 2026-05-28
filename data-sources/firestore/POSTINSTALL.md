# Llull Search Sync - Installed

Your Firestore documents are now being synced to your Llull search engine.

## Usage

Documents in your configured collection are automatically indexed. To search, make requests to your engine:

```
GET https://your-engine-url/v1/search?q=caballero&page=1&hits_per_page=10
```

### Parameters
- `q` - Search query (required)
- `page` - Page number (default: 1)
- `hits_per_page` - Results per page (default: 10)
- `use_weight` - Enable business weight ranking (`true`/`false`)
- `weight_impact` - Weight impact percentage (0.0 to 1.0)
- `fuzzy` - Enable fuzzy/typo tolerance (`true`/`false`)

### Monitoring
Check the Cloud Functions logs in the Firebase Console to monitor sync activity.
