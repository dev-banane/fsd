import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";

const API_TARGET = process.env.WEBEYE_API ?? "http://127.0.0.1:8080";
const TITLE = process.env.WEBEYE_TITLE || process.env.VITE_WEBEYE_TITLE || "WebEye";

export default defineConfig({
  plugins: [
    vue(),
    tailwindcss(),
    {
      name: "webeye-title",
      transformIndexHtml(html) {
        return html.replaceAll("__WEBEYE_TITLE__", TITLE);
      },
    },
  ],
  build: {
    outDir: "../static",
    emptyOutDir: true,
    rollupOptions: {
      output: {
        manualChunks: { maplibre: ["maplibre-gl", "@maplibre/maplibre-gl-leaflet"] },
      },
    },
  },
  server: {
    port: 5173,
    proxy: {
      "/fsd-status": { target: API_TARGET, changeOrigin: true },
      "/healthz": { target: API_TARGET, changeOrigin: true },
    },
  },
});
