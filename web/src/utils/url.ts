import type { Heartbeat } from "../types";
import { withBasePath } from "./basePath";

/** buildHeartbeatURL constructs a clickable URL for a heartbeat entry. */
export function buildHeartbeatURL(hb: Heartbeat, siteUrl?: string): string {
  const base = siteUrl || window.location.origin;
  if (hb.url) {
    try {
      return new URL(hb.url, base).toString();
    } catch {
      return hb.url;
    }
  }
  const path = withBasePath(`/api/heartbeat/${hb.id}`);
  try {
    return new URL(path, base).toString();
  } catch {
    return `${base}${path}`;
  }
}
