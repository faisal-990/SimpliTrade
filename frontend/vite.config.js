import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";
import tailwindcss from "@tailwindcss/vite";
import path from "path";

export default defineConfig({
    plugins: [react(), tailwindcss()],
    resolve: {
        alias: {
            "@": path.resolve(__dirname, "./src"),
        },
    },
    server: {
        // Proxy API calls to the Go server in dev so the SPA uses same-origin
        // "/api/..." paths (no CORS, no hardcoded host).
        proxy: {
            "/api": { target: "http://localhost:8080", changeOrigin: true },
        },
    },
});
