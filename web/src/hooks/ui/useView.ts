import { useEffect, useState } from "react";

/** VIEWS enumerates the supported dashboard screens. */
const VIEWS = ["heartbeats", "receivers", "history"] as const;
/** View represents one of the dashboard screens. */
type View = (typeof VIEWS)[number];

/** getInitialView determines the start view from the URL hash. */
function getInitialView(): View {
  const hash = window.location.hash.replace("#", "");
  if (VIEWS.includes(hash as View)) return hash as View;
  return "heartbeats";
}

/** useView manages the current dashboard view and keeps the hash in sync. */
export function useView() {
  const [view, setView] = useState<View>(() => getInitialView());

  useEffect(() => {
    const onHashChange = () => setView(getInitialView());
    window.addEventListener("hashchange", onHashChange);
    return () => window.removeEventListener("hashchange", onHashChange);
  }, []);

  return {
    view,
    updateView: (next: View) => {
      setView(next);
      window.location.hash = next;
    },
  };
}

export type { View };
