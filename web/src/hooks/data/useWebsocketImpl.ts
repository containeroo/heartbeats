import { useEffect, useState } from "react";
import type { Dispatch, SetStateAction } from "react";
import type { Heartbeat, HistoryEvent, Receiver } from "../../types";
import { withBasePath } from "../../utils/basePath";

/** useWebsocketUpdates keeps the heartbeats, receivers, and history in sync via WS. */
export function useWebsocketUpdates(params: {
  setHeartbeats: Dispatch<SetStateAction<Heartbeat[]>>;
  setReceivers: Dispatch<SetStateAction<Receiver[]>>;
  setHistory: Dispatch<SetStateAction<HistoryEvent[]>>;
}) {
  const { setHeartbeats, setReceivers, setHistory } = params;
  const [connected, setConnected] = useState(false);

  useEffect(() => {
    let alive = true;
    let socket: WebSocket | null = null;
    let retry = 0;
    const delays = [0, 200, 400, 800, 1600, 3200];

    /** applyHistoryEvent enriches state with the latest history record. */
    const applyHistoryEvent = (item: HistoryEvent) => {
      setHistory((prev) => [item, ...prev].slice(0, 500));

      setHeartbeats((prev) =>
        prev.map((hb) => {
          if (!item.heartbeatId || hb.id !== item.heartbeatId) return hb;
          const next: Heartbeat = { ...hb, hasHistory: true };
          const fields = (item.fields || {}) as Record<string, unknown>;
          if (item.type === "heartbeat_received") {
            next.lastBump = item.timestamp;
            next.status = "ok";
          }
          if (item.type === "heartbeat_transition") {
            const to = typeof fields.to === "string" ? fields.to : "";
            if (to) {
              next.status = to;
            }
          }
          return next;
        }),
      );

      if (
        item.type === "notification_delivered" ||
        item.type === "notification_failed"
      ) {
        const fields = (item.fields || {}) as Record<string, unknown>;
        const receiverId = item.receiver || "";
        const receiverType = item.targetType || "";
        const receiverTarget =
          typeof fields.target === "string" ? fields.target : "";
        const errorMsg = item.message || "";

        if (receiverId && receiverType && receiverTarget) {
          setReceivers((prev) => {
            const updated = prev.map((rv) => {
              if (
                rv.id === receiverId &&
                rv.type === receiverType &&
                rv.destination === receiverTarget
              ) {
                return {
                  ...rv,
                  lastSent: item.timestamp,
                  lastErr:
                    item.type === "notification_failed" ? errorMsg : null,
                };
              }
              return rv;
            });

            const exists = updated.some(
              (rv) =>
                rv.id === receiverId &&
                rv.type === receiverType &&
                rv.destination === receiverTarget,
            );

            if (!exists) {
              updated.push({
                id: receiverId,
                type: receiverType,
                destination: receiverTarget,
                lastSent: item.timestamp,
                lastErr: item.type === "notification_failed" ? errorMsg : null,
              });
            }

            return updated;
          });
        }
      }
    };

    /** applyHeartbeatUpdate merges incoming heartbeat snapshots. */
    const applyHeartbeatUpdate = (item: Heartbeat) => {
      if (!item?.id) return;
      setHeartbeats((prev) => {
        const idx = prev.findIndex((hb) => hb.id === item.id);
        if (idx === -1) return [...prev, item];
        const next = [...prev];
        next[idx] = { ...next[idx], ...item };
        return next;
      });
    };

    /** applyReceiverUpdate merges incoming receiver snapshots. */
    const applyReceiverUpdate = (item: Receiver) => {
      if (!item?.id || !item.type || !item.destination) return;
      setReceivers((prev) => {
        const idx = prev.findIndex(
          (rv) =>
            rv.id === item.id &&
            rv.type === item.type &&
            rv.destination === item.destination,
        );
        if (idx === -1) return [...prev, item];
        const next = [...prev];
        next[idx] = { ...next[idx], ...item };
        return next;
      });
    };

    /** connect establishes the websocket and wires event handlers. */
    const connect = () => {
      if (!alive) return;
      const protocol = window.location.protocol === "https:" ? "wss" : "ws";
      const wsURL = `${protocol}://${window.location.host}${withBasePath("/api/ws")}`;

      if (process.env.NODE_ENV !== "production") {
        console.debug("[ws] connecting to", wsURL);
      }

      socket = new WebSocket(wsURL);
      socket.onopen = () => {
        retry = 0;
        setConnected(true);
        console.debug("[ws] connected");
      };
      socket.onerror = () => {
        console.warn("[ws] error, will retry");
        // Keep status until close to avoid flicker on transient errors.
      };
      socket.onmessage = (event) => {
        try {
          const msg = JSON.parse(event.data) as {
            type: string;
            data:
              | HistoryEvent
              | Heartbeat
              | Receiver
              | {
                  heartbeats?: Heartbeat[];
                  receivers?: Receiver[];
                  history?: HistoryEvent[];
                };
          };
          if (msg.type === "history") {
            const item = msg.data as HistoryEvent;
            if (!item) return;
            applyHistoryEvent(item);
            return;
          }
          if (msg.type === "heartbeat") {
            applyHeartbeatUpdate(msg.data as Heartbeat);
            return;
          }
          if (msg.type === "receiver") {
            applyReceiverUpdate(msg.data as Receiver);
            return;
          }
          if (msg.type === "snapshot") {
            const snapshot = msg.data as {
              heartbeats?: Heartbeat[];
              receivers?: Receiver[];
              history?: HistoryEvent[];
            };
            if (snapshot.heartbeats) setHeartbeats(snapshot.heartbeats);
            if (snapshot.receivers) setReceivers(snapshot.receivers);
            if (snapshot.history) setHistory(snapshot.history);
          }
        } catch {
          // ignore bad payloads
        }
      };
      socket.onclose = () => {
        if (!alive) return;
        setConnected(false);
        retry = Math.min(retry + 1, delays.length - 1);
        const delay = delays[retry];
        console.debug("[ws] closed, retry", retry, "delay", delay);
        window.setTimeout(connect, delay);
      };
    };

    connect();

    return () => {
      alive = false;
      setConnected(false);
      if (socket) socket.close();
    };
  }, [setHeartbeats, setReceivers, setHistory]);

  return { connected };
}
