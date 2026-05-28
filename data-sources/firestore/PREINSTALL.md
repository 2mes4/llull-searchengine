# Muntaner Search Sync

## Before installing this extension

You need a Muntaner search engine instance running and accessible via HTTP. You can deploy one on any VPS or cloud provider.

**Requirements:**
- A running Muntaner search engine instance
- The engine's URL (e.g., `https://search.example.com`)
- A shared authentication token

**What this extension does:**
- Watches a Firestore collection for document changes (create, update, delete)
- Sends only the fields you specify to the Muntaner engine
- Automatically removes documents from the index when deleted from Firestore
- Supports weighted ranking via a configurable numeric field

**Firestore costs:**
This extension uses Firestore triggers which may incur costs. See [Firebase pricing](https://firebase.google.com/pricing) for details.
