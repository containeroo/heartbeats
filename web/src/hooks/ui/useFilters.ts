import { useState } from "react";

/** useFilters tracks the search queries for heartbeats, receivers, and history. */
export function useFilters() {
  const [heartbeatQuery, setHeartbeatQuery] = useState("");
  const [receiverQuery, setReceiverQuery] = useState("");
  const [historyQuery, setHistoryQuery] = useState("");

  return {
    heartbeatQuery,
    receiverQuery,
    historyQuery,
    setHeartbeatQuery,
    setReceiverQuery,
    setHistoryQuery,
  };
}
