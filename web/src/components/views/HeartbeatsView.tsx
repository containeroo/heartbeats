import { useMemo } from "react";
import type { Heartbeat } from "../../types";
import { formatDateTime, formatRecentTimestamp } from "../../utils/format";
import { heartbeatStatusClass, heartbeatStatusLabel } from "../../utils/status";
import { useHeartbeatSelection } from "../../hooks/ui/useHeartbeatSelection";
import { Modal } from "../shared/Modal";
import { HeartbeatDetails } from "../shared/HeartbeatDetails";
import { useRecentTimer } from "../../hooks/ui/useRecentTimestamp";

/** HeartbeatsView displays the list of heartbeat cards and modal detail. */
export function HeartbeatsView({
  items,
  query,
  onQueryChange,
  onSelectReceiver,
  onSelectHistory,
  siteURL,
}: {
  items: Heartbeat[];
  query: string;
  onQueryChange: (value: string) => void;
  onSelectReceiver: (id: string) => void;
  onSelectHistory: (id: string) => void;
  siteURL?: string;
}) {
  const { selected, resolvedURL, open, close } = useHeartbeatSelection(
    items,
    siteURL,
  );
  const handleReceiverClick = (receiver: string) => {
    onSelectReceiver(receiver);
    close();
  };

  const tick = useRecentTimer(items.map((hb) => hb.lastBump));

  const filtered = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return items;
    return items.filter((hb) => hb.id.toLowerCase().includes(needle));
  }, [items, query]);

  /** renderLastBump picks the right formatting for the "Last bump" column. */
  const renderLastBump = (value?: string) => {
    void tick;
    return formatRecentTimestamp(value);
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

      <div className="table table-heartbeats compact">
        <div className="table-head">
          <div>Status</div>
          <div>ID</div>
          <div>Last bump</div>
          <div>Receivers</div>
          <div>History</div>
        </div>
        <div className="table-body">
          {filtered.map((hb) => (
            <button
              className="table-row row-button"
              key={hb.id}
              type="button"
              onClick={() => open(hb.id)}
            >
              <div>
                <span className={heartbeatStatusClass(hb.status)}>
                  {heartbeatStatusLabel(hb.status)}
                </span>
              </div>
              <div>{hb.id}</div>
              <div title={hb.lastBump ? formatDateTime(hb.lastBump) : "never"}>
                {renderLastBump(hb.lastBump)}
              </div>
              <div className="tags">
                {(hb.receivers || []).slice(0, 3).map((receiver) => (
                  <button
                    key={receiver}
                    className="tag"
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      onSelectReceiver(receiver);
                    }}
                  >
                    {receiver}
                  </button>
                ))}
                {(hb.receivers || []).length > 3 ? (
                  <span className="tag muted">
                    +{(hb.receivers || []).length - 3}
                  </span>
                ) : null}
              </div>
              <div>
                {hb.hasHistory ? (
                  <button
                    className="tag ghost"
                    type="button"
                    onClick={(event) => {
                      event.stopPropagation();
                      onSelectHistory(hb.id);
                    }}
                  >
                    history
                  </button>
                ) : (
                  <span className="subtle">—</span>
                )}
              </div>
            </button>
          ))}
        </div>
      </div>

      <Modal
        open={Boolean(selected)}
        title={selected ? `Heartbeat: ${selected.id}` : "Heartbeat details"}
        onClose={close}
      >
        {selected ? (
          <HeartbeatDetails
            hb={selected}
            siteURL={resolvedURL}
            onReceiverClick={handleReceiverClick}
          />
        ) : null}
      </Modal>
    </section>
  );
}
