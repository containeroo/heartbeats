import type { View } from "../../hooks/ui/useView";

/** Topbar shows the current view name, realtime indicator, and refresh control. */
export function Topbar({
  view,
  wsConnected,
  onRefresh,
  refreshing,
}: {
  view: View;
  wsConnected: boolean;
  onRefresh: () => void;
  refreshing: boolean;
}) {
  return (
    <header className="topbar">
      <div className="topbar-left">
        <h2 className="page-title">{view}</h2>
      </div>
      <div className="topbar-actions">
        <span className={`signal-indicator ${wsConnected ? "live" : "down"}`}>
          <svg className="signal-svg" viewBox="0 0 120 16" aria-hidden="true">
            <path
              className="signal-trace"
              d="M0 8 H26 L32 8 L36 3 L42 13 L48 2 L54 8 H120"
            />
          </svg>
          <span className="signal-tooltip">
            {wsConnected ? "Realtime online" : "Realtime offline"}
          </span>
        </span>
        <button
          className={`refresh-button ${refreshing ? "refreshing" : ""}`}
          type="button"
          onClick={onRefresh}
          disabled={refreshing}
          aria-label="refresh data"
        >
          <svg
            className="refresh-icon"
            xmlns="http://www.w3.org/2000/svg"
            aria-hidden="true"
            x="0px"
            y="0px"
            width="100"
            height="100"
            viewBox="0 0 30 30"
          >
            <path d="M 15 3 C 12.031398 3 9.3028202 4.0834384 7.2070312 5.875 A 1.0001 1.0001 0 1 0 8.5058594 7.3945312 C 10.25407 5.9000929 12.516602 5 15 5 C 20.19656 5 24.450989 8.9379267 24.951172 14 L 22 14 L 26 20 L 30 14 L 26.949219 14 C 26.437925 7.8516588 21.277839 3 15 3 z M 4 10 L 0 16 L 3.0507812 16 C 3.562075 22.148341 8.7221607 27 15 27 C 17.968602 27 20.69718 25.916562 22.792969 24.125 A 1.0001 1.0001 0 1 0 21.494141 22.605469 C 19.74593 24.099907 17.483398 25 15 25 C 9.80344 25 5.5490109 21.062074 5.0488281 16 L 8 16 L 4 10 z"></path>
          </svg>
        </button>
      </div>
    </header>
  );
}
