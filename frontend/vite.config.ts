import react from "@vitejs/plugin-react";
import { defineConfig } from "vitest/config";

export default defineConfig({
  plugins: [react()],
  server: {
    // The backend sends no CORS headers, so the dev server forwards API calls (D31).
    proxy: {
      "/api": "http://127.0.0.1:8080",
    },
  },
  test: {
    environment: "jsdom",
    setupFiles: ["src/test/setup.ts"],
    coverage: {
      provider: "v8",
      include: ["src/**"],
      exclude: [
        "src/main.tsx",
        "src/test/**",
        "src/mocks/**",
        "src/**/*.test.{ts,tsx}",
      ],
      reporter: ["text", "text-summary"],
      thresholds: {
        statements: 90,
      },
    },
  },
});
