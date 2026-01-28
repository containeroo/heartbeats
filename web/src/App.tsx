import { AppShell } from "./components/shell/AppShell";
import { HeartbeatsView } from "./components/views/HeartbeatsView";
import { HistoryView } from "./components/views/HistoryView";
import { ReceiversView } from "./components/views/ReceiversView";
import { useConfig } from "./hooks/data/useConfig";
import { useDashboardData } from "./hooks/data/useDashboardData";
import { useFilters } from "./hooks/ui/useFilters";
import { useFooterText } from "./hooks/ui/useFooterText";
import { useToast } from "./hooks/ui/useToast";
import { useCallback, useEffect } from "react";
import { useView } from "./hooks/ui/useView";
import { useWebsocketUpdates } from "./hooks/data/useWebsocket";

/** App is the SPA root that wires hooks, toasts, and views. */
export default function App() {
  const { toast, show: showToast } = useToast();
  const {
    heartbeatQuery,
    receiverQuery,
    historyQuery,
    setHeartbeatQuery,
    setReceiverQuery,
    setHistoryQuery,
  } = useFilters();

  const { data: runtime } = useConfig();
  const { view, updateView } = useView();
  const {
    heartbeats,
    receivers,
    history,
    setHeartbeats,
    setReceivers,
    setHistory,
    heartbeatsLoading,
    receiversLoading,
    historyLoading,
    heartbeatsError,
    receiversError,
    historyError,
    refreshHeartbeats,
    refreshReceivers,
    refreshHistory,
    summary,
    reload,
  } = useDashboardData();

  const { connected: wsConnected } = useWebsocketUpdates({
    setHeartbeats,
    setReceivers,
    setHistory,
  });

  useEffect(() => {
    if (heartbeatsError) {
      showToast(heartbeatsError);
    }
  }, [heartbeatsError, showToast]);

  useEffect(() => {
    if (receiversError) {
      showToast(receiversError);
    }
  }, [receiversError, showToast]);

  useEffect(() => {
    if (historyError) {
      showToast(historyError);
    }
  }, [historyError, showToast]);

  async function handleReload() {
    try {
      await reload();
      showToast("Reload completed");
    } catch (error) {
      showToast(error instanceof Error ? error.message : "Reload failed");
    }
  }

  const footerText = useFooterText(runtime);

  const isRefreshing = heartbeatsLoading || receiversLoading || historyLoading;

  const refreshAll = useCallback(async () => {
    await Promise.all([
      refreshHeartbeats(),
      refreshReceivers(),
      refreshHistory(),
    ]);
  }, [refreshHeartbeats, refreshReceivers, refreshHistory]);

  const handleRefreshAll = useCallback(async () => {
    try {
      await refreshAll();
      showToast("Refresh completed");
    } catch (error) {
      showToast(error instanceof Error ? error.message : "Refresh failed");
    }
  }, [refreshAll, showToast]);

  return (
    <AppShell
      view={view}
      summary={summary}
      wsConnected={wsConnected}
      onNavigate={updateView}
      onReload={handleReload}
      toast={toast}
      footerText={footerText}
      isRefreshing={isRefreshing}
      onRefresh={handleRefreshAll}
    >
      {view === "heartbeats" ? (
        <HeartbeatsView
          items={heartbeats}
          query={heartbeatQuery}
          onQueryChange={setHeartbeatQuery}
          onSelectReceiver={(id) => {
            setReceiverQuery(id);
            updateView("receivers");
          }}
          onSelectHistory={(id) => {
            setHistoryQuery(id);
            updateView("history");
          }}
          siteURL={runtime?.siteUrl}
        />
      ) : null}
      {view === "receivers" ? (
        <ReceiversView
          items={receivers}
          query={receiverQuery}
          onQueryChange={setReceiverQuery}
        />
      ) : null}
      {view === "history" ? (
        <HistoryView
          items={history}
          query={historyQuery}
          onQueryChange={setHistoryQuery}
        />
      ) : null}
    </AppShell>
  );
}
