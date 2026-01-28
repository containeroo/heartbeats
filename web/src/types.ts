/** Heartbeat models the data needed for heartbeat cards. */
export type Heartbeat = {
  id: string;
  status: string;
  description?: string;
  interval?: string;
  intervalSeconds?: number;
  lateAfter?: string;
  lateAfterSeconds?: number;
  lastBump?: string;
  url?: string;
  receivers?: string[];
  hasHistory?: boolean;
};

/** Receiver models the receiver summary shown in the UI. */
export type Receiver = {
  id: string;
  type: string;
  destination: string;
  lastSent?: string;
  lastErr?: string | null;
};

/** HistoryEvent mirrors the backend history event payload. */
export type HistoryEvent = {
  timestamp: string;
  type: string;
  heartbeatId?: string;
  receiver?: string;
  targetType?: string;
  status?: string;
  message?: string;
  fields?: Record<string, unknown>;
};
