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
