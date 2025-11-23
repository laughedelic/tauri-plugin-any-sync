import { invoke } from "@tauri-apps/api/core";

/**
 * Ping the Go backend to test connectivity
 * @param value Optional message to send to the backend
 * @returns Promise resolving to the echoed response or null
 */
export async function ping(value?: string): Promise<string | null> {
  try {
    console.log("[any-sync] Calling ping with value:", value);
    const response = await invoke<{ value?: string }>("plugin:any-sync|ping", {
      payload: {
        value,
      },
    });
    console.log("[any-sync] Ping response:", response);
    return response.value || null;
  } catch (error) {
    console.error("[any-sync] Ping failed:", error);
    // Re-throw with more context
    throw new Error(`Failed to ping backend: ${error}`);
  }
}

/**
 * Store a document in the specified collection
 * @param collection Collection name to store the document in
 * @param id Unique identifier for the document
 * @param document Document data as a JavaScript object (will be JSON-serialized)
 * @returns Promise resolving when the document is stored
 * @throws Error if the document cannot be serialized or stored
 *
 * @example
 * ```typescript
 * await storagePut("users", "user123", { name: "Alice", age: 30 });
 * ```
 */
export async function storagePut(
  collection: string,
  id: string,
  document: any,
): Promise<void> {
  try {
    console.log(`[any-sync] Storing document in ${collection}/${id}`);

    // Serialize document to JSON string
    const documentJson = JSON.stringify(document);

    await invoke<{ success: boolean }>("plugin:any-sync|storage_put", {
      payload: {
        collection,
        id,
        documentJson,
      },
    });

    console.log(`[any-sync] Document stored successfully: ${collection}/${id}`);
  } catch (error) {
    console.error(
      `[any-sync] Storage put failed for ${collection}/${id}:`,
      error,
    );
    throw new Error(`Failed to store document: ${error}`);
  }
}

/**
 * Retrieve a document from the specified collection by ID
 * @param collection Collection name to retrieve from
 * @param id Document identifier
 * @returns Promise resolving to the document object, or null if not found
 * @throws Error if the document cannot be retrieved or parsed
 *
 * @example
 * ```typescript
 * const user = await storageGet("users", "user123");
 * if (user) {
 *   console.log(user.name); // "Alice"
 * }
 * ```
 */
export async function storageGet(
  collection: string,
  id: string,
): Promise<any | null> {
  try {
    console.log(`[any-sync] Getting document from ${collection}/${id}`);

    const response = await invoke<{ documentJson?: string }>(
      "plugin:any-sync|storage_get",
      {
        payload: {
          collection,
          id,
        },
      },
    );

    if (!response.documentJson) {
      console.log(`[any-sync] Document not found: ${collection}/${id}`);
      return null;
    }

    // Parse JSON string back to object
    const document = JSON.parse(response.documentJson);
    console.log(
      `[any-sync] Document retrieved successfully: ${collection}/${id}`,
    );
    return document;
  } catch (error) {
    console.error(
      `[any-sync] Storage get failed for ${collection}/${id}:`,
      error,
    );
    throw new Error(`Failed to retrieve document: ${error}`);
  }
}

/**
 * List all document IDs in the specified collection
 * @param collection Collection name to list documents from
 * @returns Promise resolving to an array of document IDs
 * @throws Error if the collection cannot be listed
 *
 * @example
 * ```typescript
 * const ids = await storageList("users");
 * console.log(`Found ${ids.length} users`);
 * // ["user123", "user456", ...]
 * ```
 */
export async function storageList(collection: string): Promise<string[]> {
  try {
    console.log(`[any-sync] Listing documents in collection: ${collection}`);

    const response = await invoke<{ ids: string[] }>(
      "plugin:any-sync|storage_list",
      {
        payload: {
          collection,
        },
      },
    );

    console.log(
      `[any-sync] Found ${response.ids.length} documents in ${collection}`,
    );
    return response.ids;
  } catch (error) {
    console.error(`[any-sync] Storage list failed for ${collection}:`, error);
    throw new Error(`Failed to list documents: ${error}`);
  }
}
