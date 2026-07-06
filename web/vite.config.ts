/// <reference types="vitest/config" />
import { defineConfig } from "vitest/config";

export default defineConfig({
  root: ".",
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  server: {
    proxy: {
      "/api": "http://127.0.0.1:8080",
    },
  },
  test: {
    environment: "jsdom",
    include: ["src/**/*.test.ts"],
    css: true,
  },
});
