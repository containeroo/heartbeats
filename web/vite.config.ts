import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

const routePrefix = process.env.VITE_ROUTE_PREFIX || "/";

export default defineConfig({
  base: routePrefix === "/" ? "/" : `${routePrefix}/`,
  plugins: [react()],
  build: {
    outDir: "dist",
    emptyOutDir: true,
    assetsDir: "assets"
  },
  publicDir: "public"
});
