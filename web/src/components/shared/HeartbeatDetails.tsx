import { useEffect, useRef, useState } from "react";
import type { Heartbeat } from "../../types";
import { formatDateTime } from "../../utils/format";
import { heartbeatStatusClass, heartbeatStatusLabel } from "../../utils/status";
import { buildHeartbeatURL } from "../../utils/url";

/** HeartbeatDetails shows metadata, receivers, and a copyable URL for a heartbeat. */
export function HeartbeatDetails({
  hb,
  siteURL,
  onReceiverClick,
}: {
  hb: Heartbeat;
  siteURL?: string;
  onReceiverClick?: (receiver: string) => void;
}) {
  const url = buildHeartbeatURL(hb, siteURL);
  const [copied, setCopied] = useState(false);
  const timeoutRef = useRef<number | null>(null);

  useEffect(() => {
    return () => {
      if (timeoutRef.current !== null) {
        window.clearTimeout(timeoutRef.current);
      }
    };
  }, []);

  const handleCopy = async () => {
    if (!url) return;
    await navigator.clipboard.writeText(url);
    setCopied(true);
    if (timeoutRef.current !== null) {
      window.clearTimeout(timeoutRef.current);
    }
    timeoutRef.current = window.setTimeout(() => setCopied(false), 800);
  };

  return (
    <div className="details-grid">
      <div>
        <p className="eyebrow label">ID</p>
        <p className="detail value">{hb.id}</p>
      </div>
      <div>
        <p className="eyebrow label">Status</p>
        <span
          className={`${heartbeatStatusClass(hb.status)}`}
          aria-label={`status ${heartbeatStatusLabel(hb.status)}`}
        >
          {heartbeatStatusLabel(hb.status)}
        </span>
      </div>
      <div>
        <p className="eyebrow label">Last Bump</p>
        <p className="detail value">
          {hb.lastBump ? formatDateTime(hb.lastBump) : "never"}
        </p>
      </div>
      <div>
        <p className="eyebrow label">Interval</p>
        <p className="detail value">{hb.interval || "—"}</p>
      </div>
      <div>
        <p className="eyebrow label">Late after</p>
        <p className="detail value">{hb.lateAfter || "—"}</p>
      </div>
      <div>
        <p className="eyebrow label">Receivers</p>
        <div className="details-tags">
          {(hb.receivers || []).map((r) => (
            <button
              key={r}
              type="button"
              className="tag"
              onClick={() => onReceiverClick?.(r)}
            >
              {r}
            </button>
          ))}
          {(!hb.receivers || hb.receivers.length === 0) && <span>—</span>}
        </div>
      </div>
      <div className="details-span">
        <p className="eyebrow label">URL</p>
        <div className="url-cell">
          <button className="url-text value" type="button" onClick={handleCopy}>
            {url}
          </button>
          <button
            className={`copy-btn ${copied ? "copied" : ""}`}
            type="button"
            onClick={handleCopy}
          >
            {copied ? "Copied" : "Copy"}
          </button>
        </div>
      </div>
    </div>
  );
}
