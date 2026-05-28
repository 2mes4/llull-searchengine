import { onDocumentWritten } from "firebase-functions/v2/firestore";
import * as logger from "firebase-functions/logger";

const ENDPOINT_URL = process.env.ENDPOINT_URL || "";
const AUTH_TOKEN = process.env.AUTH_TOKEN || "";
const INDEXABLE_FIELDS = (process.env.INDEXABLE_FIELDS || "").split(",").map((f) => f.trim()).filter(Boolean);
const WEIGHT_FIELD = process.env.WEIGHT_FIELD || "";
const INDEX_NAME = process.env.INDEX_NAME || "";

interface SyncPayload {
  id: string;
  action: "INDEX" | "DELETE";
  fields?: Record<string, unknown>;
  updated_at?: number;
}

function getIndexUrl(): string {
  const index = INDEX_NAME;
  if (!index) {
    return ENDPOINT_URL;
  }
  return ENDPOINT_URL.replace(/\/v1\/[^/]+\/index$/, `/v1/${index}/index`);
}

export const syncToLlull = onDocumentWritten(
  {
    document: `${process.env.COLLECTION_PATH || "documents"}/{documentId}`,
    region: process.env.DATABASE_REGION || "nam5",
  },
  async (event) => {
    const docId = event.params.documentId;
    const change = event.data;

    if (!change?.after.exists) {
      logger.info(`Document deleted: ${docId}`);
      await sendToEngine({ id: docId, action: "DELETE" });
      return;
    }

    const data = change.after.data();
    const fields: Record<string, unknown> = {};

    for (const fieldName of INDEXABLE_FIELDS) {
      if (data[fieldName] !== undefined) {
        fields[fieldName] = data[fieldName];
      }
    }

    if (WEIGHT_FIELD && data[WEIGHT_FIELD] !== undefined) {
      fields["weight"] = Number(data[WEIGHT_FIELD]) || 0;
    }

    if (Object.keys(fields).length === 0) {
      logger.warn(`No indexable fields found for document ${docId}`);
      return;
    }

    logger.info(`Syncing document ${docId} with ${Object.keys(fields).length} fields`);

    await sendToEngine({
      id: docId,
      action: "INDEX",
      fields,
      updated_at: Date.now(),
    });
  }
);

async function sendToEngine(payload: SyncPayload): Promise<void> {
  const url = getIndexUrl();

  try {
    const response = await fetch(url, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${AUTH_TOKEN}`,
      },
      body: JSON.stringify(payload),
      signal: AbortSignal.timeout(5000),
    });

    if (!response.ok) {
      logger.error(`Engine returned ${response.status} for doc ${payload.id}`);
    } else {
      logger.info(`Successfully synced ${payload.id} (${payload.action})`);
    }
  } catch (error) {
    logger.error(`Failed to sync ${payload.id}:`, error);
  }
}
