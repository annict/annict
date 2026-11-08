import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { initializeCopyableInputs } from "./copyable-input";

const VALUE = "ABCD EFGH IJKL";
const SUCCESS_MESSAGE = "手入力用のキーをコピーしました";
const FAILURE_MESSAGE = "コピーできませんでした";

// components/copyable_input.templが描画する欄。
function markup(): string {
  return `
    <div role="group" class="field" data-copyable-input>
      <label class="label" for="secret">手入力用のキー</label>
      <div class="flex gap-2">
        <input class="input" data-copyable-input-value id="secret" readonly type="text" value="${VALUE}"/>
        <button
          type="button"
          hidden
          aria-label="手入力用のキーをコピー"
          data-copyable-input-button
          data-success-message="${SUCCESS_MESSAGE}"
          data-failure-message="${FAILURE_MESSAGE}"
        >
          <svg aria-hidden="true"></svg>
        </button>
      </div>
      <p class="sr-only" role="status" data-copyable-input-status></p>
    </div>
  `;
}

function input(): HTMLInputElement {
  return document.querySelector("[data-copyable-input-value]") as HTMLInputElement;
}

function button(): HTMLButtonElement {
  return document.querySelector("[data-copyable-input-button]") as HTMLButtonElement;
}

function status(): HTMLElement {
  return document.querySelector("[data-copyable-input-status]") as HTMLElement;
}

function selectedText(element: HTMLInputElement): string {
  return element.value.slice(element.selectionStart ?? 0, element.selectionEnd ?? 0);
}

// season-filter.test.tsがWindow.basecoatを必須として宣言するため、差し替えと削除は独立した型で行う。
function basecoatHolder(): { basecoat?: { init: (componentName: string) => void } } {
  return window as unknown as { basecoat?: { init: (componentName: string) => void } };
}

function stubClipboard(writeText: (text: string) => Promise<void>): ReturnType<typeof vi.fn> {
  const mock = vi.fn(writeText);
  Object.defineProperty(navigator, "clipboard", { value: { writeText: mock }, configurable: true });
  return mock;
}

// Basecoatのトーストの代わりに、渡された設定を記録して要素を返す `#toaster` を置く。
function stubToaster(): ReturnType<typeof vi.fn> {
  const toaster = document.createElement("div") as HTMLElement & { toast?: unknown };
  toaster.id = "toaster";
  const toast = vi.fn(() => {
    const element = document.createElement("div");
    element.className = "toast";
    element.setAttribute("role", "status");
    element.setAttribute("aria-atomic", "true");
    toaster.appendChild(element);
    return element;
  });
  toaster.toast = toast;
  document.body.appendChild(toaster);
  return toast;
}

describe("initializeCopyableInputs", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    document.body.innerHTML = markup();
    initializeCopyableInputs();
  });

  afterEach(() => {
    vi.useRealTimers();
    delete basecoatHolder().basecoat;
  });

  // コピーはJSが担うため、ボタンはJSが動くまで隠して描画される。
  it("初期化でコピーボタンを表示する", () => {
    expect(button().hidden).toBe(false);
  });

  // 初期化を繰り返しても、1回のクリックで書き込むのは1回だけにする。
  it("初期化を繰り返してもリスナーを重ねない", async () => {
    const writeText = stubClipboard(async () => {});
    stubToaster();
    initializeCopyableInputs();

    button().click();
    await vi.runAllTimersAsync();

    expect(writeText).toHaveBeenCalledTimes(1);
  });

  // 値は空白で区切っていて、ダブルクリックやタップでは一部しか選ばれないため、欄に入ったら全体を選ぶ。
  it("キーボードでフォーカスしたら値の全体を選ぶ", () => {
    input().focus();

    expect(selectedText(input())).toBe(VALUE);
  });

  // マウスでは、フォーカスで選んだ後にmouseupでカーソルが置かれるため、クリックでも選び直す。
  it("クリックしたら値の全体を選ぶ", () => {
    input().setSelectionRange(0, 0);

    input().click();

    expect(selectedText(input())).toBe(VALUE);
  });

  it("ボタンでクリップボードに値を書き、トーストとステータス領域で知らせる", async () => {
    const writeText = stubClipboard(async () => {});
    const toast = stubToaster();

    button().click();
    await vi.runAllTimersAsync();

    expect(writeText).toHaveBeenCalledWith(VALUE);
    expect(toast).toHaveBeenCalledWith({ category: "success", description: SUCCESS_MESSAGE });
    expect(status().textContent).toBe(SUCCESS_MESSAGE);
  });

  // 読み上げはステータス領域が受け持つため、トーストは支援技術に二重に読み上げさせない。
  it("トーストからライブリージョンのロールを外す", async () => {
    stubClipboard(async () => {});
    stubToaster();

    button().click();
    await vi.runAllTimersAsync();

    const toastElement = document.querySelector("#toaster .toast") as HTMLElement;
    expect(toastElement.hasAttribute("role")).toBe(false);
    expect(toastElement.hasAttribute("aria-atomic")).toBe(false);
  });

  // 同じ文言を入れ直しても読み上げられるよう、いったん空にしてから入れる。
  it("続けてコピーしたらステータス領域の文言を空にしてから入れ直す", async () => {
    stubClipboard(async () => {});
    stubToaster();

    button().click();
    await vi.runAllTimersAsync();
    button().click();
    await vi.advanceTimersByTimeAsync(0);

    expect(status().textContent).toBe("");

    await vi.runAllTimersAsync();

    expect(status().textContent).toBe(SUCCESS_MESSAGE);
  });

  it("クリップボードに書けなかったら失敗を知らせ、値の全体を選ぶ", async () => {
    stubClipboard(async () => {
      throw new Error("拒否された");
    });
    const toast = stubToaster();

    button().click();
    await vi.runAllTimersAsync();

    expect(toast).toHaveBeenCalledWith({ category: "error", description: FAILURE_MESSAGE });
    expect(status().textContent).toBe(FAILURE_MESSAGE);
    expect(document.activeElement).toBe(input());
    expect(selectedText(input())).toBe(VALUE);
  });

  // Basecoatが文言をテキストとしてエスケープするため、二重にエスケープしないよう文言をそのまま渡す。
  it("トーストの文言をエスケープせずに渡す", async () => {
    stubClipboard(async () => {});
    const toast = stubToaster();
    button().dataset.successMessage = "<b>コピー</b>&";

    button().click();
    await vi.runAllTimersAsync();

    expect(toast).toHaveBeenCalledWith({ category: "success", description: "<b>コピー</b>&" });
  });

  // フラッシュメッセージが無い画面には `#toaster` が描画されないため、作ってBasecoatに初期化させる。
  it("#toasterが無ければ作ってBasecoatに初期化させる", async () => {
    stubClipboard(async () => {});
    const init = vi.fn();
    basecoatHolder().basecoat = { init };

    button().click();
    await vi.runAllTimersAsync();

    const toaster = document.getElementById("toaster");
    expect(toaster?.className).toBe("toaster");
    expect(init).toHaveBeenCalledWith("toaster");
  });
});
