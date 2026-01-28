import { useEffect, useState } from "react";
import { getConfig } from "../../api";

/** RuntimeConfig mirrors the /api/config payload. */
type RuntimeConfig = {
  version: string;
  commit: string;
  siteUrl: string;
};

/** useConfig fetches runtime metadata at startup. */
export function useConfig() {
  const [data, setData] = useState<RuntimeConfig | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let active = true;
    getConfig()
      .then((resp) => {
        if (!active) return;
        setData(resp);
      })
      .catch((err: unknown) => {
        if (!active) return;
        setError(err instanceof Error ? err.message : "Failed to load config");
      });
    return () => {
      active = false;
    };
  }, []);

  return { data, error };
}
