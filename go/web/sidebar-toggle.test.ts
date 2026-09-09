import { beforeEach, describe, expect, it, vi } from "vitest";

import { initializeSidebarToggle } from "./sidebar-toggle";

// Build the minimal DOM the toggle wiring expects: a trigger that references the
// sidebar by id (data-sidebar-toggle), and a sidebar exposing a spy-able
// toggle() — Basecoat's imperative open/close API. aria-hidden seeds the
// sidebar's initial open state.
//
// [Ja] トグル結線が前提とする最小の DOM を組み立てる。id で sidebar を参照する
// トリガー (data-sidebar-toggle) と、スパイ可能な toggle() を持つ sidebar
// (Basecoat の命令的な開閉 API)。aria-hidden で sidebar の初期の開閉状態を仕込む。
function setupDom(
  ariaHidden: "true" | "false",
  desktopOpen: "true" | "false" = ariaHidden === "false" ? "true" : "false",
) {
  document.body.innerHTML = `
    <button data-sidebar-toggle="db-sidebar" aria-expanded="true"></button>
    <aside id="db-sidebar" class="sidebar" data-desktop-open="${desktopOpen}" aria-hidden="${ariaHidden}"></aside>
  `;
  const trigger = document.querySelector("[data-sidebar-toggle]") as HTMLButtonElement;
  const sidebar = document.getElementById("db-sidebar") as HTMLElement & {
    close: () => void;
    open: () => void;
    toggle: () => void;
  };
  sidebar.close = vi.fn(() => sidebar.setAttribute("aria-hidden", "true"));
  sidebar.open = vi.fn(() => sidebar.setAttribute("aria-hidden", "false"));
  sidebar.toggle = vi.fn(() =>
    sidebar.setAttribute("aria-hidden", sidebar.getAttribute("aria-hidden") === "false" ? "true" : "false"),
  );
  return { trigger, sidebar };
}

// Build the db.templ layout shape the mobile-overlay logic depends on: the skip
// link before the sidebar, a sidebar with a focusable child, the content wrapper
// as its next sibling (holding the trigger, mirroring how the toggle now lives in
// each page's title row), and a spy-able toggle(). aria-hidden seeds the open
// state; width seeds the viewport so isMobileSidebar() resolves against the 768px
// default breakpoint.
//
// [Ja] モバイルオーバーレイのロジックが依存する db.templ のレイアウト構造を組み立てる。
// sidebar より前のスキップリンク、フォーカス可能な子を持つ sidebar、その次の兄弟である
// content ラッパー (トグルが各ページのタイトル行へ移った現状に合わせ、トリガーを内包する)、
// スパイ可能な toggle()。aria-hidden で開閉状態を、width でビューポートを仕込み、
// isMobileSidebar() を既定の 768px ブレークポイントに対して解決させる。
function setupLayout(options: { ariaHidden: "true" | "false"; desktopOpen?: "true" | "false"; width: number }) {
  setViewportWidth(options.width);
  document.body.innerHTML = `
    <a href="#db-main" id="db-skip-link">skip</a>
    <aside id="db-sidebar" class="sidebar" data-side="left" data-desktop-open="${options.desktopOpen ?? (options.ariaHidden === "false" ? "true" : "false")}" aria-hidden="${options.ariaHidden}">
      <button data-sidebar-close="db-sidebar" id="db-sidebar-close">close</button>
      <a href="/db" id="db-sidebar-link">nav</a>
    </aside>
    <div id="db-content">
      <main id="db-main" tabindex="-1">
        <button data-sidebar-toggle="db-sidebar" aria-controls="db-sidebar" aria-expanded="true">toggle</button>
        <input id="db-content-input"/>
      </main>
    </div>
  `;
  const trigger = document.querySelector("[data-sidebar-toggle]") as HTMLButtonElement;
  const sidebar = document.getElementById("db-sidebar") as HTMLElement & {
    close: () => void;
    open: () => void;
    toggle: () => void;
  };
  const content = sidebar.nextElementSibling as HTMLElement;
  const skipLink = document.getElementById("db-skip-link") as HTMLAnchorElement;
  const closeButton = document.getElementById("db-sidebar-close") as HTMLButtonElement;
  const sidebarLink = document.getElementById("db-sidebar-link") as HTMLElement;
  const contentInput = document.getElementById("db-content-input") as HTMLInputElement;
  sidebar.close = vi.fn(() => sidebar.setAttribute("aria-hidden", "true"));
  sidebar.open = vi.fn(() => sidebar.setAttribute("aria-hidden", "false"));
  sidebar.toggle = vi.fn(() =>
    sidebar.setAttribute("aria-hidden", sidebar.getAttribute("aria-hidden") === "false" ? "true" : "false"),
  );
  return { trigger, sidebar, content, skipLink, closeButton, sidebarLink, contentInput };
}

// window.innerWidth is read-only in the DOM lib types, so cast to assign the
// simulated viewport width that isMobileSidebar() reads.
//
// [Ja] window.innerWidth は DOM lib の型では読み取り専用のため、isMobileSidebar()
// が読むビューポート幅を差し込むにはキャストして代入する。
function setViewportWidth(width: number): void {
  (window as unknown as { innerWidth: number }).innerWidth = width;
}

// Let queued MutationObserver callbacks run: happy-dom (like browsers) delivers
// them on a microtask, so awaiting a resolved promise flushes them.
//
// [Ja] キューされた MutationObserver コールバックを実行させる。happy-dom は
// (ブラウザと同様) それらを microtask で配送するため、解決済み Promise を await
// すれば flush できる。
const flushObservers = () => Promise.resolve();

describe("initializeSidebarToggle", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    window.history.replaceState(null, "", "/db/works");
    document.cookie = "annict_db_sidebar_open=; Path=/db; Max-Age=0; SameSite=Lax; Secure";
    setViewportWidth(1024);
  });

  it("syncs a trigger's aria-expanded from the sidebar's aria-hidden on init", () => {
    setViewportWidth(500);
    const { trigger } = setupDom("true");

    initializeSidebarToggle();

    expect(trigger.getAttribute("aria-expanded")).toBe("false");
  });

  it("calls the sidebar's toggle() when a trigger is clicked", () => {
    const { trigger, sidebar } = setupDom("false");

    initializeSidebarToggle();
    trigger.click();

    expect(sidebar.toggle).toHaveBeenCalledTimes(1);
  });

  it("syncs aria-expanded when the sidebar closes via a non-trigger path", async () => {
    const { trigger, sidebar } = setupLayout({ ariaHidden: "false", width: 1024 });

    initializeSidebarToggle();
    expect(trigger.getAttribute("aria-expanded")).toBe("true");

    // Basecoat closes the sidebar itself (e.g. overlay click) by flipping
    // aria-hidden; the observer must catch that path too.
    //
    // [Ja] Basecoat はオーバーレイクリックなどでサイドバー自身を閉じる際に
    // aria-hidden を切り替える。observer はその経路も捕捉しなければならない。
    sidebar.setAttribute("aria-hidden", "true");
    await flushObservers();

    expect(trigger.getAttribute("aria-expanded")).toBe("false");
  });

  it("makes the background inert and moves focus into the sidebar when opened on mobile", async () => {
    const { trigger, sidebar, content, closeButton } = setupLayout({
      ariaHidden: "true",
      width: 500,
    });

    initializeSidebarToggle();
    expect(content.inert).toBe(false);

    sidebar.setAttribute("aria-hidden", "false");
    await flushObservers();

    expect(content.inert).toBe(true);
    expect(document.activeElement).toBe(closeButton);
    expect(trigger.getAttribute("aria-expanded")).toBe("true");
  });

  it("makes elements before the sidebar inert too while the mobile overlay is open", async () => {
    const { sidebar, skipLink } = setupLayout({ ariaHidden: "true", width: 500 });

    initializeSidebarToggle();
    expect(skipLink.inert).toBe(false);

    // The skip link sits before the sidebar in the DOM, so it is not covered by the
    // content wrapper's inert state; it has to be inerted as part of the background.
    //
    // [Ja] スキップリンクは DOM 上でサイドバーより前にあるため content ラッパーの inert
    // では覆われない。背面の一部として inert にする必要がある。
    sidebar.setAttribute("aria-hidden", "false");
    await flushObservers();
    expect(skipLink.inert).toBe(true);

    sidebar.setAttribute("aria-hidden", "true");
    await flushObservers();
    expect(skipLink.inert).toBe(false);
  });

  it("keeps elements before the sidebar interactive on desktop", () => {
    const { skipLink } = setupLayout({ ariaHidden: "false", width: 1024 });

    initializeSidebarToggle();

    expect(skipLink.inert).toBe(false);
  });

  it("clears inert and returns focus to the toggle when closed on mobile", async () => {
    const { trigger, sidebar, content } = setupLayout({ ariaHidden: "true", width: 500 });

    initializeSidebarToggle();

    // Click records the toggle as the return target before the overlay opens, so
    // it is captured while the trigger is still interactive (not yet inert).
    //
    // [Ja] クリックはオーバーレイが開く前にトグルを復帰先として記録する。トリガーが
    // まだ操作可能 (inert 化前) のうちに捕捉するため。
    trigger.click();

    sidebar.setAttribute("aria-hidden", "false");
    await flushObservers();
    expect(content.inert).toBe(true);

    sidebar.setAttribute("aria-hidden", "true");
    await flushObservers();

    expect(content.inert).toBe(false);
    expect(document.activeElement).toBe(trigger);
    expect(trigger.getAttribute("aria-expanded")).toBe("false");
  });

  it("closes the mobile sidebar with Escape and restores the page state", async () => {
    const { trigger, sidebar, content, closeButton } = setupLayout({
      ariaHidden: "true",
      width: 500,
    });

    initializeSidebarToggle();
    trigger.click();
    sidebar.setAttribute("aria-hidden", "false");
    await flushObservers();

    expect(content.inert).toBe(true);
    expect(document.activeElement).toBe(closeButton);

    closeButton.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    await flushObservers();

    expect(sidebar.close).toHaveBeenCalledTimes(1);
    expect(sidebar.getAttribute("aria-hidden")).toBe("true");
    expect(trigger.getAttribute("aria-expanded")).toBe("false");
    expect(content.inert).toBe(false);
    expect(document.activeElement).toBe(trigger);
  });

  it("does not close the mobile sidebar when Escape was already handled", async () => {
    const { trigger, sidebar, sidebarLink } = setupLayout({ ariaHidden: "true", width: 500 });

    initializeSidebarToggle();
    trigger.click();
    sidebar.setAttribute("aria-hidden", "false");
    await flushObservers();

    const event = new KeyboardEvent("keydown", {
      key: "Escape",
      bubbles: true,
      cancelable: true,
    });
    event.preventDefault();
    sidebarLink.dispatchEvent(event);

    expect(sidebar.close).not.toHaveBeenCalled();
    expect(sidebar.getAttribute("aria-hidden")).toBe("false");
  });

  it("uses the server-rendered desktop state and persists changes in a Cookie", async () => {
    const { trigger, sidebar, closeButton } = setupLayout({
      ariaHidden: "true",
      desktopOpen: "false",
      width: 1024,
    });

    initializeSidebarToggle();

    expect(sidebar.close).not.toHaveBeenCalled();
    expect(sidebar.getAttribute("aria-hidden")).toBe("true");
    expect(trigger.getAttribute("aria-expanded")).toBe("false");

    trigger.click();
    await flushObservers();
    expect(sidebar.dataset.desktopOpen).toBe("true");
    expect(document.cookie).toContain("annict_db_sidebar_open=true");

    closeButton.click();
    await flushObservers();
    expect(sidebar.dataset.desktopOpen).toBe("false");
    expect(document.cookie).toContain("annict_db_sidebar_open=false");
    expect(document.activeElement).toBe(trigger);
  });

  it("keeps the mobile sidebar closed even when the desktop preference is open", () => {
    const { sidebar } = setupLayout({ ariaHidden: "true", desktopOpen: "true", width: 500 });

    initializeSidebarToggle();

    expect(sidebar.open).not.toHaveBeenCalled();
    expect(sidebar.getAttribute("aria-hidden")).toBe("true");
  });

  it("restores the desktop preference after crossing the breakpoint", () => {
    const { sidebar } = setupLayout({ ariaHidden: "true", desktopOpen: "true", width: 500 });

    initializeSidebarToggle();
    setViewportWidth(1024);
    window.dispatchEvent(new Event("resize"));

    expect(sidebar.open).toHaveBeenCalledTimes(1);
    expect(sidebar.getAttribute("aria-hidden")).toBe("false");
  });

  it("closes the desktop sidebar when crossing into mobile", () => {
    const { sidebar } = setupLayout({ ariaHidden: "false", width: 1024 });

    initializeSidebarToggle();
    setViewportWidth(500);
    window.dispatchEvent(new Event("resize"));

    expect(sidebar.close).toHaveBeenCalledTimes(1);
    expect(sidebar.getAttribute("aria-hidden")).toBe("true");
  });

  it("does not change an open sidebar during a resize within the mobile breakpoint", () => {
    const { sidebar } = setupLayout({ ariaHidden: "false", width: 500 });

    initializeSidebarToggle();
    setViewportWidth(600);
    window.dispatchEvent(new Event("resize"));

    expect(sidebar.close).not.toHaveBeenCalled();
    expect(sidebar.getAttribute("aria-hidden")).toBe("false");
  });

  it("keeps the background interactive on desktop even while the sidebar is open", () => {
    const { sidebar, content } = setupLayout({ ariaHidden: "false", width: 1024 });

    initializeSidebarToggle();

    expect(content.inert).toBe(false);
    expect(sidebar.contains(document.activeElement)).toBe(false);
  });

  it("clears the mobile inert state when the viewport widens to desktop", () => {
    const { content } = setupLayout({ ariaHidden: "false", width: 500 });

    initializeSidebarToggle();
    expect(content.inert).toBe(true);

    setViewportWidth(1024);
    window.dispatchEvent(new Event("resize"));

    expect(content.inert).toBe(false);
  });

  // Open then close once on mobile so a trigger is recorded and focus has already
  // returned to it, then move focus into the (now interactive) content. This is the
  // starting state for the "no focus steal on re-sync" cases below.
  //
  // [Ja] モバイルで一度開いてから閉じ、トグルが記録されフォーカスがそこへ戻った状態に
  // してから、(操作可能になった) コンテンツへフォーカスを移す。下の「再同期でフォーカスを
  // 奪わない」ケースの初期状態。
  async function focusContentAfterMobileClose() {
    const layout = setupLayout({ ariaHidden: "true", width: 500 });

    initializeSidebarToggle();
    layout.trigger.click();
    layout.sidebar.setAttribute("aria-hidden", "false");
    await flushObservers();
    layout.sidebar.setAttribute("aria-hidden", "true");
    await flushObservers();
    expect(document.activeElement).toBe(layout.trigger);

    layout.contentInput.focus();
    expect(document.activeElement).toBe(layout.contentInput);
    return layout;
  }

  it("does not steal focus back to the toggle on a mobile-width resize while closed", async () => {
    const { contentInput } = await focusContentAfterMobileClose();

    setViewportWidth(600);
    window.dispatchEvent(new Event("resize"));

    expect(document.activeElement).toBe(contentInput);
  });

  it("does not steal focus back to the toggle when a node is added to the body while closed", async () => {
    const { contentInput } = await focusContentAfterMobileClose();

    // A node inserted anywhere under <body> (e.g. a flash toast) triggers a full re-sync.
    //
    // [Ja] <body> 配下のどこかへノードが追加される (例: flash トースト) と全体が再同期される。
    document.body.append(document.createElement("div"));
    await flushObservers();

    expect(document.activeElement).toBe(contentInput);
  });
});
