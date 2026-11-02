// コピーするための値の欄 (components/copyable_input.templ) に振る舞いを付ける。
//
// 認証アプリの手入力用のキーは書き写しやすいよう4文字ごとに空白で区切っているため、ダブルクリック
// やタップでは区切りの1組しか選ばれない。欄は値をコピーするためだけにあるので、クリックやタップ、
// キーボードでフォーカスしたら値の全体を選ぶ。横のボタンはクリップボードへ値を書き、結果をトーストと
// ステータス領域で知らせる。イベントは文書で受け、欄の有無や描画の経路に関わらず同じ振る舞いにする。

const CONTAINER_SELECTOR = "[data-copyable-input]";
const VALUE_SELECTOR = "[data-copyable-input-value]";
const BUTTON_SELECTOR = "[data-copyable-input-button]";
const STATUS_SELECTOR = "[data-copyable-input-status]";

// ステータス領域の文言を空にしてから入れ直すまでの間。同じ文言を続けて入れても、いったん空に
// したことが支援技術に伝わってから入れ直さないと、2回目のコピーが読み上げられない。
const ANNOUNCE_DELAY_MS = 100;

type ToastCategory = "success" | "error";

// Basecoatのトースト (basecoat-css/all) が `#toaster` に生やすメソッド。
interface Toaster extends HTMLElement {
  toast?: (config: { category: ToastCategory; description: string }) => HTMLElement;
}

interface BasecoatRuntime {
  init: (componentName: string) => void;
}

let listening = false;

export function initializeCopyableInputs(): void {
  // ボタンはJSが動くまで隠して描画される (components/copyable_input.templ)。
  for (const button of document.querySelectorAll<HTMLButtonElement>(BUTTON_SELECTOR)) {
    button.hidden = false;
  }

  // リスナーは文書に付くため、何度呼ばれても1組だけにする。
  if (listening) return;
  listening = true;

  document.addEventListener("focusin", handleSelect);
  // マウスでは、フォーカスした後のmouseupでカーソルが置かれて選択が外れるため、clickでも選び直す。
  document.addEventListener("click", handleSelect);
  document.addEventListener("click", handleCopyClick);
}

function handleSelect(event: Event): void {
  const target = event.target;
  if (!(target instanceof HTMLInputElement) || !target.matches(VALUE_SELECTOR)) return;

  target.select();
}

function handleCopyClick(event: Event): void {
  const target = event.target;
  if (!(target instanceof Element)) return;

  const button = target.closest<HTMLButtonElement>(BUTTON_SELECTOR);
  if (!button) return;

  void copy(button);
}

async function copy(button: HTMLButtonElement): Promise<void> {
  const container = button.closest(CONTAINER_SELECTOR);
  const input = container?.querySelector<HTMLInputElement>(VALUE_SELECTOR);
  if (!container || !input) return;

  // 安全でない文脈ではnavigator.clipboardが無く、権限が無ければwriteTextが拒否される。どちらも
  // 失敗として伝え、利用者が自分でコピーできるよう値を選んでおく。
  try {
    await navigator.clipboard.writeText(input.value);
  } catch {
    input.focus();
    input.select();
    notify(container, button.dataset.failureMessage ?? "", "error");
    return;
  }

  notify(container, button.dataset.successMessage ?? "", "success");
}

function notify(container: Element, message: string, category: ToastCategory): void {
  if (message === "") return;

  announce(container, message);
  showToast(message, category);
}

// トーストは表示のたびに要素を作るため、作った要素のロールで読み上げを期待すると支援技術によっては
// 伝わらない。読み上げは描画時からあるステータス領域に任せる。
function announce(container: Element, message: string): void {
  const status = container.querySelector(STATUS_SELECTOR);
  if (!status) return;

  status.textContent = "";
  window.setTimeout(() => {
    status.textContent = message;
  }, ANNOUNCE_DELAY_MS);
}

function showToast(message: string, category: ToastCategory): void {
  const toaster = findOrCreateToaster();
  if (typeof toaster.toast !== "function") return;

  const toast = toaster.toast({ category, description: escapeHTML(message) });

  // 読み上げはステータス領域が受け持つため、トーストのロールを外して二重に読み上げさせない。
  toast.removeAttribute("role");
  toast.removeAttribute("aria-atomic");
}

// `#toaster` はフラッシュメッセージがあるときだけサーバーが描画する (components/flash.templ)。
// 無ければ作り、Basecoatに初期化させる。
function findOrCreateToaster(): Toaster {
  const existing = document.getElementById("toaster");
  if (existing) return existing as Toaster;

  const toaster = document.createElement("div");
  toaster.id = "toaster";
  toaster.className = "toaster";
  document.body.appendChild(toaster);

  const basecoat = (window as Window & { basecoat?: BasecoatRuntime }).basecoat;
  basecoat?.init("toaster");

  return toaster as Toaster;
}

// Basecoatはdescriptionをそのままマークアップとして差し込むため、文言をエスケープして渡す。
function escapeHTML(text: string): string {
  const element = document.createElement("span");
  element.textContent = text;
  return element.innerHTML;
}
