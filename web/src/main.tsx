import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./styles.css";

declare global {
  interface Window {
    __HEARTBEATS_BASE?: string;
  }
}

/** readRoutePrefix extracts the optional base path from the meta tag. */
const readRoutePrefix = (): string => {
  const meta = document.querySelector('meta[name="routePrefix"]');
  // Guard: return empty base when the meta tag is missing.
  if (!meta) return "";
  const value = meta.getAttribute("content")?.trim() ?? "";
  // Guard: ignore default or templated values.
  if (!value || value === "/" || value.includes("{{")) return "";
  return value.endsWith("/") ? value.slice(0, -1) : value;
};

// Expose the base path so helpers can prefix asset URLs.
window.__HEARTBEATS_BASE = readRoutePrefix();

// Render the SPA into the root node.
ReactDOM.createRoot(document.getElementById("root")!).render(
  <React.StrictMode>
    <App />
  </React.StrictMode>,
);
