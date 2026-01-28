declare global {
  interface Window {
    __HEARTBEATS_BASE?: string;
  }
}

/** getBasePath returns the currently configured base path for the SPA. */
const getBasePath = (): string => {
  const raw = window.__HEARTBEATS_BASE ?? "";
  // Normalize "" and "/" to an empty prefix.
  if (!raw || raw === "/") return "";
  return raw.endsWith("/") ? raw.slice(0, -1) : raw;
};

/** withBasePath prefixes a path with the current base path. */
export const withBasePath = (path: string): string => {
  const base = getBasePath();
  // Keep absolute paths intact when no base is configured.
  if (!base) return path.startsWith("/") ? path : `/${path}`;
  if (path.startsWith("/")) return `${base}${path}`;
  return `${base}/${path}`;
};

/** resolveMediaUrl returns a URL suitable for media assets, honoring the base path. */
export const resolveMediaUrl = (url?: string): string => {
  if (!url) return "";
  // Allow absolute or blob/data URLs without modification.
  if (/^(https?:|blob:|data:)/.test(url)) return url;
  return withBasePath(url);
};
