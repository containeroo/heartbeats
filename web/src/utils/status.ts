/** heartbeatStatusLabel returns a normalized label for heartbeat states. */
export function heartbeatStatusLabel(status?: string): string {
  if (!status) return "unknown";
  return status;
}

/** heartbeatStatusClass returns CSS classes for heartbeat badges. */
export function heartbeatStatusClass(status?: string): string {
  switch (status) {
    case "ok":
      return "status-pill status-ok";
    case "missing":
      return "status-pill status-missing";
    case "late":
      return "status-pill status-late";
    case "never":
      return "status-pill status-never";
    default:
      return "status-pill status-unknown";
  }
}

/** receiverStatusClass returns CSS classes for receiver badges. */
export function receiverStatusClass(
  lastSent?: string,
  lastErr?: string | null,
): string {
  if (!lastSent) return "status-pill status-missing";
  if (!lastErr) return "status-pill status-ok";
  return "status-pill status-unknown";
}

/** receiverStatusLabel returns a readable label for receiver status. */
export function receiverStatusLabel(
  lastSent?: string,
  lastErr?: string | null,
): string {
  if (!lastSent) return "never";
  if (!lastErr) return "ok";
  return "error";
}
