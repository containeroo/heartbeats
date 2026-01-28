import { useCallback, useEffect, useMemo, useState } from "react";
import {
  listHeartbeats,
  listHistory,
  listReceivers,
  reloadConfig,
} from "../../api";
import type { Heartbeat, HistoryEvent, Receiver } from "../../types";

/** useDashboardData centralizes the data loading and state for the dashboard. */
export function useDashboardData() {
  const [heartbeats, setHeartbeats] = useState<Heartbeat[]>([]);
  const [receivers, setReceivers] = useState<Receiver[]>([]);
  const [history, setHistory] = useState<HistoryEvent[]>([]);

  const [heartbeatsLoading, setHeartbeatsLoading] = useState(false);
  const [receiversLoading, setReceiversLoading] = useState(false);
  const [historyLoading, setHistoryLoading] = useState(false);

  const [heartbeatsError, setHeartbeatsError] = useState<string | null>(null);
  const [receiversError, setReceiversError] = useState<string | null>(null);
  const [historyError, setHistoryError] = useState<string | null>(null);

  const refreshHeartbeats = useCallback(async () => {
    setHeartbeatsLoading(true);
    setHeartbeatsError(null);
    try {
      const data = await listHeartbeats();
      setHeartbeats(data ?? []);
    } catch (error) {
      setHeartbeatsError(
        error instanceof Error ? error.message : "Failed to load",
      );
    } finally {
      setHeartbeatsLoading(false);
    }
  }, []);

  const refreshReceivers = useCallback(async () => {
    setReceiversLoading(true);
    setReceiversError(null);
    try {
      const data = await listReceivers();
      setReceivers(data ?? []);
    } catch (error) {
      setReceiversError(
        error instanceof Error ? error.message : "Failed to load",
      );
    } finally {
      setReceiversLoading(false);
    }
  }, []);

  const refreshHistory = useCallback(async () => {
    setHistoryLoading(true);
    setHistoryError(null);
    try {
      const data = await listHistory();
      setHistory(data ?? []);
    } catch (error) {
      setHistoryError(
        error instanceof Error ? error.message : "Failed to load",
      );
    } finally {
      setHistoryLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshHeartbeats();
    refreshReceivers();
    refreshHistory();
  }, [refreshHeartbeats, refreshReceivers, refreshHistory]);

  const summary = useMemo(() => {
    const ok = heartbeats.filter((hb) => hb.status === "ok").length;
    const missing = heartbeats.filter((hb) => hb.status === "missing").length;
    const late = heartbeats.filter((hb) => hb.status === "late").length;
    return { ok, missing, late, total: heartbeats.length };
  }, [heartbeats]);

  const reload = useCallback(async () => {
    await reloadConfig();
  }, []);

  return {
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
  };
}
