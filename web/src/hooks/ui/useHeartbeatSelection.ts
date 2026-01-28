import { useMemo, useState } from "react";
import type { Heartbeat } from "../../types";
import { buildHeartbeatURL } from "../../utils/url";

/** useHeartbeatSelection manages the modal selection state for the heartbeat list. */
export function useHeartbeatSelection(items: Heartbeat[], siteUrl?: string) {
  const [selectedID, setSelectedID] = useState<string | null>(null);

  const selected = useMemo(() => {
    if (!selectedID) return null;
    return items.find((hb) => hb.id === selectedID) || null;
  }, [items, selectedID]);

  const resolvedURL = useMemo(() => {
    if (!selected) return "";
    return buildHeartbeatURL(selected, siteUrl);
  }, [selected, siteUrl]);

  return {
    selected,
    resolvedURL,
    open: (id: string) => setSelectedID(id),
    close: () => setSelectedID(null),
  };
}
