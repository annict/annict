import { beforeEach, describe, expect, it, vi } from "vitest";

import { initializeSidebarToggle } from "./sidebar-toggle";

// トグル結線が前提とする最小のDOMを組み立てる。idでsidebarを参照する
// トリガー (data-sidebar-toggle) と、スパイ可能なtoggle() を持つsidebar
// (Basecoatの命令的な開閉API)。aria-hiddenでsidebarの初期の開閉状態を仕込む。
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

// モバイルオーバーレイのロジックが依存するdb.templのレイアウト構造を組み立てる。
// sidebarより前のスキップリンク、フォーカス可能な子を持つsidebar、その次の兄弟である
// contentラッパー (トグルが各ページのタイトル行へ移った現状に合わせ、トリガーを内包する)、
// スパイ可能なtoggle()。aria-hiddenで開閉状態を、widthでビューポートを仕込み、
// isMobileSidebar() を既定の768pxブレークポイントに対して解決させる。
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

// window.innerWidthはDOM libの型では読み取り専用のため、isMobileSidebar()
// が読むビューポート幅を差し込むにはキャストして代入する。
function setViewportWidth(width: number): void {
  (window as unknown as { innerWidth: number }).innerWidth = width;
}

// キューされたMutationObserverコールバックを実行させる。happy-domは
// (ブラウザと同様) それらをmicrotaskで配送するため、解決済みPromiseをawait
// すればflushできる。
const flushObservers = () => Promise.resolve();

describe("initializeSidebarToggle", () => {
  beforeEach(() => {
    document.body.innerHTML = "";
    window.history.replaceState(null, "", "/db/works");
    document.cookie = "annict_db_sidebar_open=; Path=/db; Max-Age=0; SameSite=Lax; Secure";
    setViewportWidth(1024);
  });

  it("初期化時にサイドバーのaria-hiddenをもとにトリガーのaria-expandedを同期する", () => {
    setViewportWidth(500);
    const { trigger } = setupDom("true");

    initializeSidebarToggle();

    expect(trigger.getAttribute("aria-expanded")).toBe("false");
  });

  it("トリガーをクリックするとサイドバーのtoggle()を呼び出す", () => {
    const { trigger, sidebar } = setupDom("false");

    initializeSidebarToggle();
    trigger.click();

    expect(sidebar.toggle).toHaveBeenCalledTimes(1);
  });

  it("トリガー以外の経路でサイドバーが閉じたときもaria-expandedを同期する", async () => {
    const { trigger, sidebar } = setupLayout({ ariaHidden: "false", width: 1024 });

    initializeSidebarToggle();
    expect(trigger.getAttribute("aria-expanded")).toBe("true");

    // Basecoatはオーバーレイクリックなどでサイドバー自身を閉じる際に
    // aria-hiddenを切り替える。observerはその経路も捕捉しなければならない。
    sidebar.setAttribute("aria-hidden", "true");
    await flushObservers();

    expect(trigger.getAttribute("aria-expanded")).toBe("false");
  });

  it("モバイルで開くと背面をinertにし、サイドバー内へフォーカスを移す", async () => {
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

  it("モバイルのオーバーレイ表示中はサイドバーより前の要素もinertにする", async () => {
    const { sidebar, skipLink } = setupLayout({ ariaHidden: "true", width: 500 });

    initializeSidebarToggle();
    expect(skipLink.inert).toBe(false);

    // スキップリンクはDOM上でサイドバーより前にあるためcontentラッパーのinert
    // では覆われない。背面の一部としてinertにする必要がある。
    sidebar.setAttribute("aria-hidden", "false");
    await flushObservers();
    expect(skipLink.inert).toBe(true);

    sidebar.setAttribute("aria-hidden", "true");
    await flushObservers();
    expect(skipLink.inert).toBe(false);
  });

  it("デスクトップではサイドバーより前の要素を操作可能なままにする", () => {
    const { skipLink } = setupLayout({ ariaHidden: "false", width: 1024 });

    initializeSidebarToggle();

    expect(skipLink.inert).toBe(false);
  });

  it("モバイルで閉じるとinertを解除し、トグルへフォーカスを戻す", async () => {
    const { trigger, sidebar, content } = setupLayout({ ariaHidden: "true", width: 500 });

    initializeSidebarToggle();

    // クリックはオーバーレイが開く前にトグルを復帰先として記録する。トリガーが
    // まだ操作可能 (inert化前) のうちに捕捉するため。
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

  it("Escapeキーでモバイルのサイドバーを閉じ、ページの状態を復元する", async () => {
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

  it("Escapeキーが処理済みの場合はモバイルのサイドバーを閉じない", async () => {
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

  it("サーバー描画時のデスクトップの開閉状態を使い、変更をCookieに保存する", async () => {
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

  it("デスクトップの設定が開いた状態でもモバイルではサイドバーを閉じたままにする", () => {
    const { sidebar } = setupLayout({ ariaHidden: "true", desktopOpen: "true", width: 500 });

    initializeSidebarToggle();

    expect(sidebar.open).not.toHaveBeenCalled();
    expect(sidebar.getAttribute("aria-hidden")).toBe("true");
  });

  it("ブレークポイントをまたぐとデスクトップの開閉設定を復元する", () => {
    const { sidebar } = setupLayout({ ariaHidden: "true", desktopOpen: "true", width: 500 });

    initializeSidebarToggle();
    setViewportWidth(1024);
    window.dispatchEvent(new Event("resize"));

    expect(sidebar.open).toHaveBeenCalledTimes(1);
    expect(sidebar.getAttribute("aria-hidden")).toBe("false");
  });

  it("デスクトップからモバイルの幅に変わるとサイドバーを閉じる", () => {
    const { sidebar } = setupLayout({ ariaHidden: "false", width: 1024 });

    initializeSidebarToggle();
    setViewportWidth(500);
    window.dispatchEvent(new Event("resize"));

    expect(sidebar.close).toHaveBeenCalledTimes(1);
    expect(sidebar.getAttribute("aria-hidden")).toBe("true");
  });

  it("モバイルの幅の範囲内でリサイズしても開いているサイドバーの状態を変えない", () => {
    const { sidebar } = setupLayout({ ariaHidden: "false", width: 500 });

    initializeSidebarToggle();
    setViewportWidth(600);
    window.dispatchEvent(new Event("resize"));

    expect(sidebar.close).not.toHaveBeenCalled();
    expect(sidebar.getAttribute("aria-hidden")).toBe("false");
  });

  it("デスクトップではサイドバーが開いていても背面を操作可能にする", () => {
    const { sidebar, content } = setupLayout({ ariaHidden: "false", width: 1024 });

    initializeSidebarToggle();

    expect(content.inert).toBe(false);
    expect(sidebar.contains(document.activeElement)).toBe(false);
  });

  it("ビューポートがデスクトップの幅に広がるとモバイルで設定したinertを解除する", () => {
    const { content } = setupLayout({ ariaHidden: "false", width: 500 });

    initializeSidebarToggle();
    expect(content.inert).toBe(true);

    setViewportWidth(1024);
    window.dispatchEvent(new Event("resize"));

    expect(content.inert).toBe(false);
  });

  // モバイルで一度開いてから閉じ、トグルが記録されフォーカスがそこへ戻った状態に
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

  it("閉じた状態でモバイルの幅の範囲内でリサイズしてもトグルへフォーカスを戻さない", async () => {
    const { contentInput } = await focusContentAfterMobileClose();

    setViewportWidth(600);
    window.dispatchEvent(new Event("resize"));

    expect(document.activeElement).toBe(contentInput);
  });

  it("閉じた状態でbodyにノードが追加されてもトグルへフォーカスを戻さない", async () => {
    const { contentInput } = await focusContentAfterMobileClose();

    // <body> 配下のどこかへノードが追加される (例: flashトースト) と全体が再同期される。
    document.body.append(document.createElement("div"));
    await flushObservers();

    expect(document.activeElement).toBe(contentInput);
  });
});
