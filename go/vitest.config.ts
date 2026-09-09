import { defineConfig } from "vitest/config";

export default defineConfig({
  test: {
    // The frontend under web/ drives the DOM (document, MutationObserver, inert,
    // focus), so tests need a DOM environment rather than bare Node. globals stay
    // off; tests import from "vitest" explicitly.
    //
    // [Ja] web/ のフロントエンドは DOM (document, MutationObserver, inert,
    // フォーカス) を操作するため、テストには素の Node ではなく DOM 環境が要る。
    // globals は使わず、テストは "vitest" から明示 import する。
    environment: "happy-dom",
    include: ["web/**/*.test.ts"],
  },
});
