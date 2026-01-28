import { useEffect, useState } from "react";

/** useRecentTimer ticks while at least one timestamp is deemed recent. */
export function useRecentTimer(
  values: Array<string | undefined>,
  thresholdMs = 15000,
) {
  const [tick, setTick] = useState(0);

  useEffect(() => {
    const now = Date.now();
    const hasRecent = values.some((value) => {
      if (!value) {
        return false;
      }
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) {
        return false;
      }
      return now - date.getTime() <= thresholdMs;
    });
    if (!hasRecent) {
      return undefined;
    }
    const id = window.setInterval(() => setTick((count) => count + 1), 1000);
    return () => window.clearInterval(id);
  }, [values, thresholdMs]);

  return tick;
}
