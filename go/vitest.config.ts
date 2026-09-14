import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    // web/ のフロントエンドはDOM (document, MutationObserver, inert,
    // フォーカス) を操作するため、テストには素のNodeではなくDOM環境が要る。
    // globalsは使わず、テストは "vitest" から明示importする。
    environment: "happy-dom",
    include: ["web/**/*.test.ts"],
  },
});
