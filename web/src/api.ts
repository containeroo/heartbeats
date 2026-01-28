import type { Heartbeat, HistoryEvent, Receiver } from "./types";
import { withBasePath } from "./utils/basePath";

/** request dispatches a JSON request to the backend and unwraps responses. */
async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(withBasePath(path), {
    headers: {
      "Content-Type": "application/json",
      ...(init?.headers || {}),
    },
    ...init,
  });

  if (!res.ok) {
    let message = res.statusText;
    try {
      const body = await res.json();
      message = body?.error || message;
    } catch {
      // ignore
    }
    throw new Error(message);
  }

  if (res.status === 204) {
    // @ts-expect-error allow void response
    return null;
  }

  return res.json() as Promise<T>;
}

/** listHeartbeats returns the available heartbeat summaries. */
export async function listHeartbeats(): Promise<Heartbeat[]> {
  return request("/api/heartbeats");
}

/** listReceivers returns the configured receiver summaries. */
export async function listReceivers(): Promise<Receiver[]> {
  return request("/api/receivers");
}

/** listHistory returns the recorded history events. */
export async function listHistory(): Promise<HistoryEvent[]> {
  return request("/api/history");
}

/** reloadConfig triggers a server-side config reload. */
export async function reloadConfig(): Promise<void> {
  return request("/-/reload", { method: "POST" });
}

/** getConfig returns runtime metadata such as commit, version, and site URL. */
export async function getConfig(): Promise<{
  version: string;
  commit: string;
  siteUrl: string;
}> {
  return request("/api/config");
}
