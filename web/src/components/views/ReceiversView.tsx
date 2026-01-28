import { useMemo } from "react";
import type { Receiver } from "../../types";
import { formatDateTime, formatRecentTimestamp } from "../../utils/format";
import { receiverStatusClass, receiverStatusLabel } from "../../utils/status";
import { useRecentTimer } from "../../hooks/ui/useRecentTimestamp";

/** ReceiversView renders the receiver table with statuses. */
export function ReceiversView({
  items,
  query,
  onQueryChange,
}: {
  items: Receiver[];
  query: string;
  onQueryChange: (value: string) => void;
}) {
  const tick = useRecentTimer(items.map((rv) => rv.lastSent));

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return items;
    return items.filter((rv) => rv.id.toLowerCase().includes(needle));
  }, [items, query]);

  return (
    <section className="panel">
      <div className="panel-controls">
        <label className="search">
          <span>Filter</span>
          <input
            value={query}
            onChange={(event) => onQueryChange(event.target.value)}
            placeholder="Search receiver ID"
            type="search"
          />
        </label>
      </div>

      <div className="table table-receivers">
        <div className="table-head">
          <div>Receiver</div>
          <div>Type</div>
          <div>Destination</div>
          <div>Last sent</div>
          <div>Status</div>
        </div>
        <div className="table-body">
          {filtered.map((rv) => (
            <div
              className="table-row"
              key={`${rv.id}-${rv.type}-${rv.destination}`}
            >
              <div>{rv.id}</div>
              <div>{rv.type}</div>
              <div>{rv.destination}</div>
              <div title={rv.lastSent ? formatDateTime(rv.lastSent) : "never"}>
                {(() => {
                  void tick;
                  return formatRecentTimestamp(rv.lastSent);
                })()}
              </div>
              <div>
                <span className={receiverStatusClass(rv.lastSent, rv.lastErr)}>
                  {receiverStatusLabel(rv.lastSent, rv.lastErr)}
                </span>
                {rv.lastErr ? <p className="subtle">{rv.lastErr}</p> : null}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}
