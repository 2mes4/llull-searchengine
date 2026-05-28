# Data Source: Firestore

Syncs changes from a [Google Cloud Firestore](https://cloud.google.com/firestore) collection into the Llull search engine using a Firebase Extension with a Cloud Function trigger.

## Overview

This data source uses a **push model**: a Cloud Function triggers on Firestore document writes and pushes changes to the Llull engine via HTTP. This provides near real-time sync with zero polling overhead.

## Prerequisites

- A Firebase project with Firestore enabled
- A running Llull engine instance accessible via HTTPS
- Firebase CLI installed (`npm install -g firebase-tools`)

## Installation

### Install from local source

```bash
cd data-sources/firestore
firebase ext:install . --project=your-project-id
```

### Manual deployment

```bash
cd data-sources/firestore/functions
npm install
npm run build
firebase deploy --only functions --project=your-project-id
```

## Configuration Parameters

| Parameter | Required | Description |
|-----------|----------|-------------|
| `COLLECTION_PATH` | Yes | Firestore collection path (e.g. `users`) |
| `DATABASE_REGION` | Yes | Firestore database region (e.g. `nam5`) |
| `ENDPOINT_URL` | Yes | Llull engine URL (e.g. `https://search.example.com/v1/index`) |
| `AUTH_TOKEN` | Yes | Shared Bearer token (stored as Cloud Secret) |
| `INDEXABLE_FIELDS` | Yes | Comma-separated fields to index (e.g. `title,content,author`) |
| `WEIGHT_FIELD` | No | Numeric field for business weight ranking |

## How It Works

1. A Firestore document is created, updated, or deleted
2. The Cloud Function triggers on `document.v1.written`
3. Only the configured `INDEXABLE_FIELDS` are extracted
4. An authenticated HTTP POST is sent to the Llull engine
5. The engine enqueues and processes the change asynchronously

## Testing Locally

```bash
cd data-sources/firestore/functions
npm install
npm run build
firebase emulators:start --only functions
```
