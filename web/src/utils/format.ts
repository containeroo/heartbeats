/** formatDateTime renders an ISO timestamp with second precision. */
export function formatDateTime(value?: string): string {
  if (!value) return "—";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  const pad = (num: number) => String(num).padStart(2, "0");
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${pad(
    date.getHours(),
  )}:${pad(date.getMinutes())}:${pad(date.getSeconds())}`;
}

/** formatRelative renders a relative time description. */
export function formatRelative(value?: string): string {
  if (!value) return "never";
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  if (date.getUTCFullYear() <= 1) return "never";
  const deltaSeconds = Math.round((date.getTime() - Date.now()) / 1000);
  const absSeconds = Math.abs(deltaSeconds);
  const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: "auto" });

  if (absSeconds < 60) return rtf.format(deltaSeconds, "second");
  const minutes = Math.round(deltaSeconds / 60);
  if (Math.abs(minutes) < 60) return rtf.format(minutes, "minute");
  const hours = Math.round(minutes / 60);
  if (Math.abs(hours) < 24) return rtf.format(hours, "hour");
  const days = Math.round(hours / 24);
  return rtf.format(days, "day");
}

/** formatDuration normalizes duration strings for display. */
export function formatDuration(value?: string): string {
  if (!value) return "—";
  return value;
}

/** formatRecentTimestamp chooses relative or absolute formatting for recent dates. */
export function formatRecentTimestamp(
  value?: string,
  now = Date.now(),
): string {
  if (!value) return "never";
  const date = new Date(value);
  if (Number.isNaN(date.getTime()) || date.getUTCFullYear() <= 1)
    return "never";
  const diffSeconds = Math.floor((now - date.getTime()) / 1000);
  if (diffSeconds <= 15) {
    return formatRelative(value);
  }
  return formatDateTime(value);
}
