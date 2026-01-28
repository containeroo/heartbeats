import { useMemo } from "react";

/** RuntimeConfig carries version and commit metadata. */
type RuntimeConfig = {
  version: string;
  commit: string;
};

/** useFooterText renders the footer line combining runtime metadata and copyright. */
export function useFooterText(runtime?: RuntimeConfig | null) {
  return useMemo(() => {
    const year = new Date().getFullYear();
    const bits = [`© ${year} Heartbeats`];
    if (runtime?.version) bits.push(`v${runtime.version}`);
    if (runtime?.commit) bits.push(runtime.commit.slice(0, 7));
    return bits.join(" • ");
  }, [runtime]);
}
