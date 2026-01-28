import { useMemo } from "react";
import type { HistoryEvent } from "../../types";
import { formatDateTime } from "../../utils/format";

/** HistoryView renders the paginated history events table. */
export function HistoryView({
  items,
  query,
  onQueryChange,
}: {
  items: HistoryEvent[];
  query: string;
  onQueryChange: (value: string) => void;
}) {
  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return items;
    return items.filter((ev) =>
      (ev.heartbeatId || "").toLowerCase().includes(needle),
    );
  }, [items, query]);

  /** formatDetails produces a user-friendly line for each history event. */
  const formatDetails = (event: HistoryEvent) => {
    const fields = (event.fields || {}) as Record<string, unknown>;
    const asText = (value: unknown): string =>
      typeof value === "string" ? value : "";
    if (
      event.message &&
      event.type !== "notification_failed" &&
      event.type !== "http_access"
    ) {
      return event.message;
    }
    switch (event.type) {
      case "notification_delivered": {
        const receiver = event.receiver || "";
        const type = event.targetType || "";
        const target = asText(fields.target);
        if (!receiver && !type && !target) return "—";
        return `Notification sent to «${receiver}» via ${type} (${target})`;
      }
      case "notification_failed": {
        const receiver = event.receiver || "";
        const type = event.targetType || "";
        const target = asText(fields.target);
        const errorMsg = event.message || "";
        if (!receiver && !type && !target) return "—";
        if (errorMsg) {
          return `Notification to «${receiver}» via ${type} (${target}) failed: ${errorMsg}`;
        }
        return `Notification to «{receiver}» via ${type} (${target}) failed`;
      }
      case "heartbeat_transition": {
        const from = asText(fields.from);
        const to = asText(fields.to);
        const since = asText(fields.since);
        if (!from && !to) return "—";
        if (since) return `${from} → ${to} (after ${since})`;
        return `${from} → ${to}`;
      }
      case "heartbeat_received": {
        const payloadBytes = fields.payload_bytes;
        const enqueued = fields.enqueued;
        const size =
          typeof payloadBytes === "number"
            ? `${payloadBytes} bytes`
            : payloadBytes
              ? String(payloadBytes)
              : "";
        const enqueueText =
          typeof enqueued === "boolean"
            ? enqueued
              ? "enqueued"
              : "direct"
            : enqueued
              ? String(enqueued)
              : "";
        if (!size && !enqueueText) return "—";
        return [size, enqueueText].filter(Boolean).join(" · ");
      }
      case "http_access": {
        const entries = Object.entries(fields || {}).sort(([a], [b]) =>
          a.localeCompare(b),
        );
        if (entries.length === 0) return "—";
        return (
          <details className="detail-toggle">
            <summary>
              <span className="detail">http access</span>
              <span className="detail-chevron" aria-hidden="true" />
            </summary>
            <div className="detail-grid">
              {entries.map(([key, value]) => (
                <div className="detail-row" key={key}>
                  <span className="detail-key">{key}</span>
                  <span className="value">
                    {typeof value === "string"
                      ? value
                      : typeof value === "number" || typeof value === "boolean"
                        ? String(value)
                        : JSON.stringify(value)}
                  </span>
                </div>
              ))}
            </div>
          </details>
        );
      }
      default:
        return "—";
    }
  };

  return (
    <section className="panel">
      <div className="panel-controls">
        <label className="search">
          <span>Filter</span>
          <input
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder="Search heartbeat ID"
            type="search"
          />
        </label>
      </div>

      <div className="table table-history">
        <div className="table-head">
          <div>Time</div>
          <div>Heartbeat</div>
          <div>Details</div>
        </div>
        <div className="table-body">
          {filtered.map((event) => (
            <div
              className="table-row"
              key={`${event.timestamp}-${event.heartbeatId || "global"}-${event.type}`}
            >
              <div className="nowrap">{formatDateTime(event.timestamp)}</div>
              <div>{event.heartbeatId || "—"}</div>
              <div>{formatDetails(event)}</div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
