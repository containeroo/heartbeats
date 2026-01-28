/** VIEWS enumerates the available top-level views. */
const VIEWS = ["heartbeats", "receivers", "history"] as const;

/** View represents a permitted navigation target. */
type View = (typeof VIEWS)[number];

/** Summary captures heartbeat health counts for the sidebar card. */
type Summary = {
  total: number;
  ok: number;
  late: number;
  missing: number;
};
/** Summary captures heartbeat health counts for the sidebar card. */

/** Sidebar renders navigation, heartbeat metrics, and the config reload action. */
export function Sidebar({
  view,
  summary,
  onNavigate,
  onReload,
}: {
  view: View;
  summary: Summary;
  onNavigate: (next: View) => void;
  onReload: () => void;
}) {
  return (
    <aside className="sidebar">
      <div className="brand">
        <div aria-hidden="true">
          <img
            className="logo-mark"
            src="/heartbeats-red.svg"
            alt="Heartbeats logo"
          />
        </div>
        <div>
          <p className="eyebrow">Heartbeats</p>
        </div>
      </div>

      <nav className="nav">
        {VIEWS.map((tab) => (
          <button
            key={tab}
            type="button"
            className={`nav-link ${view === tab ? "active" : ""}`}
            onClick={() => onNavigate(tab)}
          >
            {tab}
          </button>
        ))}
      </nav>

      <section className="summary-card">
        <p className="eyebrow">Heartbeat Health</p>
        <div className="summary-grid">
          <div className="metric">
            <span>Total</span>
            <strong>{summary.total}</strong>
          </div>
          <div className="metric ok">
            <span>OK</span>
            <strong>{summary.ok}</strong>
          </div>
          <div className="metric late">
            <span>Late</span>
            <strong>{summary.late}</strong>
          </div>
          <div className="metric missing">
            <span>Missing</span>
            <strong>{summary.missing}</strong>
          </div>
        </div>
      </section>

      <button className="btn ghost" onClick={onReload} type="button">
        Reload config
      </button>
    </aside>
  );
}

export type { View };
