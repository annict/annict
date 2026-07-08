// Generic toggle wiring for the Basecoat sidebar component. A trigger marked with
// data-sidebar-toggle="<id>" toggles the matching <aside class="sidebar" id="<id>"> and keeps its
// own aria-expanded in sync with the sidebar's state. Event delegation on the document means any
// number of triggers (and triggers added to the DOM later) work without re-binding.
//
// Basecoat's sidebar runtime owns the open/closed state as aria-hidden (and inert) on the <aside>
// and exposes an imperative toggle() method (basecoat-css >= 1.0). It emits no per-toggle event and
// does not manage aria-expanded on an external trigger, so this module drives .toggle() on click and
// mirrors the resulting aria-hidden onto every trigger's aria-expanded.
//
// [Ja] Basecoat サイドバーコンポーネント用の汎用トグル結線。data-sidebar-toggle="<id>" を持つ
// トリガーが、対応する <aside class="sidebar" id="<id>"> をトグルし、自身の aria-expanded を
// サイドバーの状態に同期させる。document へのイベント委譲を使うため、トリガーが何個あっても
// (後から DOM に追加されても) 再バインド不要で動作する。
//
// Basecoat のサイドバーランタイムは開閉状態を <aside> の aria-hidden (と inert) として管理し、
// 命令的な toggle() メソッドを公開する (basecoat-css >= 1.0)。toggle 時のイベント発行も、外部
// トリガーの aria-expanded の管理もしないため、本モジュールがクリックで .toggle() を呼び、結果の
// aria-hidden を各トリガーの aria-expanded に反映する。

interface SidebarElement extends HTMLElement {
  toggle?: () => void;
}

export function initializeSidebarToggle(): void {
  document.addEventListener("click", handleClick);

  // Basecoat initializes sidebars on DOMContentLoaded, before this runs (main.js imports
  // basecoat-css first), so the sidebar's aria-hidden already reflects the initial open state
  // here. Sync once so aria-expanded is correct even when a sidebar starts collapsed (e.g. mobile).
  //
  // [Ja] Basecoat はサイドバーを DOMContentLoaded で、この処理より前に初期化する (main.js が
  // basecoat-css を先に import するため)。よってこの時点でサイドバーの aria-hidden は初期の開閉
  // 状態を反映済み。サイドバーが畳んだ状態で始まる場合 (例: モバイル) でも aria-expanded を正しく
  // 保つため、一度同期する。
  syncAllTriggers();
}

function handleClick(event: MouseEvent): void {
  const target = event.target as Element | null;
  const trigger = target?.closest<HTMLElement>("[data-sidebar-toggle]");
  if (!trigger) {
    return;
  }

  const sidebar = getSidebar(trigger);
  if (!sidebar) {
    return;
  }

  sidebar.toggle?.();
  syncTrigger(trigger, sidebar);
}

function syncAllTriggers(): void {
  document.querySelectorAll<HTMLElement>("[data-sidebar-toggle]").forEach((trigger) => {
    const sidebar = getSidebar(trigger);
    if (sidebar) {
      syncTrigger(trigger, sidebar);
    }
  });
}

function syncTrigger(trigger: HTMLElement, sidebar: HTMLElement): void {
  const open = sidebar.getAttribute("aria-hidden") === "false";
  trigger.setAttribute("aria-expanded", String(open));
}

function getSidebar(trigger: HTMLElement): SidebarElement | null {
  const id = trigger.getAttribute("data-sidebar-toggle");
  return id ? (document.getElementById(id) as SidebarElement | null) : null;
}
