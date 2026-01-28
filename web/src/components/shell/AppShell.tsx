import type { ReactNode } from "react";
import type { View } from "../../hooks/ui/useView";
import { Sidebar } from "./Sidebar";
import { Topbar } from "./Topbar";

/** Summary contains the aggregate heartbeat metrics shown in the sidebar. */
type Summary = {
  total: number;
  ok: number;
  late: number;
  missing: number;
};

/** AppShell orchestrates the sidebar, top bar, and main view container. */
export function AppShell({
  view,
  summary,
  wsConnected,
  onNavigate,
  onReload,
  onRefresh,
  isRefreshing,
  toast,
  footerText,
  children,
}: {
  view: View;
  summary: Summary;
  wsConnected: boolean;
  onNavigate: (next: View) => void;
  onReload: () => void;
  onRefresh: () => void;
  isRefreshing: boolean;
  toast: string | null;
  footerText: string;
  children: ReactNode;
}) {
  return (
    <div className="app">
      <Sidebar
        view={view}
        summary={summary}
        onNavigate={onNavigate}
        onReload={onReload}
      />

      <main className="content">
        <Topbar
          view={view}
          wsConnected={wsConnected}
          onRefresh={onRefresh}
          refreshing={isRefreshing}
        />
        {children}
      </main>

      {toast ? <div className="toast">{toast}</div> : null}

      <footer className="app-footer">
        <div className="app-footer-inner">{footerText}</div>
      </footer>
    </div>
  );
}
