// Basecoatサイドバーコンポーネント用の汎用トグル結線。data-sidebar-toggle="<id>" を持つ
// トリガーが、対応する <aside class="sidebar" id="<id>"> をトグルし、同じサイドバーを参照する
// 全トリガーを状態に同期させる。documentへのイベント委譲とDOM監視を使うため、トリガーが
// 何個あっても (後からDOMに追加されても) 再バインド不要で動作する。
//
// Basecoatのサイドバーランタイムは開閉状態を <aside> のaria-hidden (とinert) として管理し、
// 命令的なopen()・close()・toggle() メソッドを公開する (basecoat-css >= 1.0)。外部トリガーの
// aria-expanded・フォーカス・Escapeキーによる閉鎖・背面のinert・デスクトップの開閉設定保存は
// 管理しないため、本モジュールがaria-hiddenを監視してBasecoatのどの閉鎖経路にも同期し、
// Cookieへのデスクトップ設定保存、モバイルオーバーレイのフォーカス・Escape閉鎖・背面のinert
// も管理する。

interface SidebarElement extends HTMLElement {
  close?: () => void;
  open?: () => void;
  toggle?: () => void;
}

const sidebarSelector = ".sidebar[id]";
const triggerSelector = "[data-sidebar-toggle]";
const closeSelector = "[data-sidebar-close]";
const desktopOpenCookieName = "annict_db_sidebar_open";
const desktopOpenCookieMaxAge = 365 * 24 * 60 * 60;
const focusableSelector =
  'a[href], button:not([disabled]), input:not([disabled]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
const lastTriggerBySidebar = new WeakMap<SidebarElement, HTMLElement>();
const openStateBySidebar = new WeakMap<SidebarElement, boolean>();
const mobileStateBySidebar = new WeakMap<SidebarElement, boolean>();

export function initializeSidebarToggle(): void {
  document.addEventListener("click", handleClick);
  document.addEventListener("keydown", handleKeydown);
  window.addEventListener("resize", syncAllSidebars);

  const observer = new MutationObserver(handleMutations);
  observer.observe(document.body, {
    attributes: true,
    attributeFilter: ["aria-hidden"],
    childList: true,
    subtree: true,
  });

  // BasecoatはサイドバーをDOMContentLoadedで、この処理より前に初期化する (main.jsが
  // basecoat-cssを先にimportするため)。よってこの時点でサイドバーのaria-hiddenは初期の開閉
  // 状態を反映済み。サイドバーが畳んだ状態で始まる場合 (例: モバイル) でもaria-expandedと
  // ページコンテンツの状態を正しく保つため、一度同期する。
  syncAllSidebars();
}

function handleClick(event: MouseEvent): void {
  const target = event.target as Element | null;
  const closeControl = target?.closest<HTMLElement>(closeSelector);
  if (closeControl) {
    const sidebar = getControlledSidebar(closeControl, "data-sidebar-close");
    if (!sidebar) {
      return;
    }

    const returnTrigger = lastTriggerBySidebar.get(sidebar) ?? getTriggers(sidebar)[0];
    if (returnTrigger) {
      lastTriggerBySidebar.set(sidebar, returnTrigger);
    }
    sidebar.close?.();
    syncSidebar(sidebar);
    return;
  }

  const trigger = target?.closest<HTMLElement>(triggerSelector);
  if (!trigger) {
    return;
  }

  const sidebar = getControlledSidebar(trigger, "data-sidebar-toggle");
  if (!sidebar) {
    return;
  }

  lastTriggerBySidebar.set(sidebar, trigger);
  sidebar.toggle?.();
  syncSidebar(sidebar);
}

function handleKeydown(event: KeyboardEvent): void {
  if (event.key !== "Escape" || event.defaultPrevented) {
    return;
  }

  const sidebar = (document.activeElement as Element | null)?.closest<SidebarElement>(sidebarSelector);
  if (sidebar?.getAttribute("aria-hidden") === "false" && isMobileSidebar(sidebar)) {
    sidebar.close?.();
  }
}

function handleMutations(records: MutationRecord[]): void {
  let hasAddedNodes = false;

  for (const record of records) {
    if (record.type === "attributes") {
      syncSidebar(record.target as SidebarElement);
    } else if (record.type === "childList" && record.addedNodes.length > 0) {
      hasAddedNodes = true;
    }
  }

  if (hasAddedNodes) {
    syncAllSidebars();
  }
}

function syncAllSidebars(): void {
  document.querySelectorAll<SidebarElement>(sidebarSelector).forEach(syncSidebar);
}

function syncSidebar(sidebar: SidebarElement): void {
  const triggers = getTriggers(sidebar);
  if (triggers.length === 0) {
    return;
  }

  syncViewportMode(sidebar, triggers);

  const open = sidebar.getAttribute("aria-hidden") === "false";
  const previousOpen = openStateBySidebar.get(sidebar);
  const wasOpen = previousOpen ?? false;
  openStateBySidebar.set(sidebar, open);

  for (const trigger of triggers) {
    trigger.setAttribute("aria-expanded", String(open));
  }

  if (!isMobileSidebar(sidebar) && previousOpen !== undefined && previousOpen !== open) {
    writeDesktopOpenPreference(sidebar, open);
  }

  const background = getBackgroundElements(sidebar);

  if (!isMobileSidebar(sidebar)) {
    setInert(background, false);
    restoreFocusAfterClose(open, wasOpen, sidebar);
    return;
  }

  if (open) {
    setInert(background, true);
    if (!sidebar.contains(document.activeElement)) {
      sidebar.querySelector<HTMLElement>(focusableSelector)?.focus();
    }
    return;
  }

  setInert(background, false);
  restoreFocusAfterClose(open, wasOpen, sidebar);
}

// getBackgroundElementsはサイドバーの兄弟をすべて返す。モバイルオーバーレイ表示中は
// それらすべてが背面に位置するため、inertにする対象になる。直後の兄弟だけでなく全兄弟を
// 取ることで、DOM上でサイドバーより前に置かれた要素 (スキップリンクなど) も対象に含まれ、
// オーバーレイ背面で操作可能なまま残らない。
function getBackgroundElements(sidebar: SidebarElement): HTMLElement[] {
  const siblings = sidebar.parentElement?.children ?? [];
  return Array.from(siblings).filter(
    (element): element is HTMLElement => element !== sidebar && element instanceof HTMLElement,
  );
}

function setInert(elements: HTMLElement[], inert: boolean): void {
  for (const element of elements) {
    element.inert = inert;
  }
}

function restoreFocusAfterClose(open: boolean, wasOpen: boolean, sidebar: SidebarElement): void {
  const lastTrigger = lastTriggerBySidebar.get(sidebar);
  // トグルへフォーカスを戻すのは開 -> 閉の遷移時だけにする。閉じたままで起きる
  // 再同期では戻さない。syncAllSidebars() はwindowのresizeや <body> 配下へのノード
  // 追加でも走るため、このガードが無いと無関係な再同期が、利用者の移したフォーカスを
  // トグルへ奪い返してしまう。
  if (!open && wasOpen && lastTrigger && document.activeElement !== lastTrigger) {
    lastTrigger.focus();
  }
}

function syncViewportMode(sidebar: SidebarElement, triggers: HTMLElement[]): void {
  const mobile = isMobileSidebar(sidebar);
  const wasMobile = mobileStateBySidebar.get(sidebar);
  mobileStateBySidebar.set(sidebar, mobile);

  if (wasMobile === undefined) {
    if (!mobile) {
      applyDesktopOpenPreference(sidebar);
    }
    return;
  }

  if (mobile === wasMobile) {
    return;
  }

  if (mobile) {
    const returnTrigger = lastTriggerBySidebar.get(sidebar) ?? triggers[0];
    if (returnTrigger) {
      lastTriggerBySidebar.set(sidebar, returnTrigger);
    }
    sidebar.close?.();
    return;
  }

  applyDesktopOpenPreference(sidebar);
}

function applyDesktopOpenPreference(sidebar: SidebarElement): void {
  const preferredOpen = sidebar.dataset.desktopOpen !== "false";
  const open = sidebar.getAttribute("aria-hidden") === "false";
  if (preferredOpen === open) {
    return;
  }

  if (preferredOpen) {
    sidebar.open?.();
  } else {
    sidebar.close?.();
  }
}

function writeDesktopOpenPreference(sidebar: SidebarElement, open: boolean): void {
  sidebar.dataset.desktopOpen = String(open);
  document.cookie = [
    `${desktopOpenCookieName}=${String(open)}`,
    "Path=/db",
    `Max-Age=${desktopOpenCookieMaxAge}`,
    "SameSite=Lax",
    "Secure",
  ].join("; ");
}

function getTriggers(sidebar: SidebarElement): HTMLElement[] {
  return Array.from(document.querySelectorAll<HTMLElement>(triggerSelector)).filter(
    (trigger) => trigger.getAttribute("data-sidebar-toggle") === sidebar.id,
  );
}

function getControlledSidebar(control: HTMLElement, attribute: string): SidebarElement | null {
  const id = control.getAttribute(attribute);
  return id ? (document.getElementById(id) as SidebarElement | null) : null;
}

function isMobileSidebar(sidebar: SidebarElement): boolean {
  const breakpoint = Number.parseInt(sidebar.dataset.breakpoint ?? "", 10) || 768;
  return window.innerWidth < breakpoint;
}
