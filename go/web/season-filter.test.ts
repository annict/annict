import "basecoat-css/basecoat";
import "basecoat-css/combobox";

import { beforeEach, describe, expect, it } from "vitest";

import { initializeSeasonFilter } from "./season-filter";

interface BasecoatRuntime {
  init(componentName: string, options?: { force?: boolean }): void;
  stop(): void;
}

declare global {
  interface Window {
    basecoat: BasecoatRuntime;
  }
}

interface BasecoatComboboxElement extends HTMLElement {
  _destroy?: () => void;
}

function basecoatRuntime(): BasecoatRuntime {
  return window.basecoat;
}

// サーバー描画されるプログレッシブエンハンスメント用DOMを組み立てる。ネイティブ
// selectは初期状態で表示されnameを持ち、Basecoatとブリッジの両方が準備できるまで
// comboboxは非表示にする。selectedはサーバーが渡す初期選択を設定する。
function setupDom(
  selected: string[] = [],
  basecoatInitialized = true,
  removeLabelTemplate = "Remove {label}",
): {
  combobox: HTMLElement;
  input: HTMLInputElement;
  label: HTMLLabelElement;
  select: HTMLSelectElement;
} {
  document.body.innerHTML = `
    <form>
      <label id="season-label" for="season-select">Release Season</label>
      <div data-season-slugs-filter>
        <select id="season-select" name="season_slugs" data-season-slugs-select multiple>
          <option value="2024-spring">2024 Spring</option>
          <option value="2024-summer">2024 Summer</option>
        </select>
        <div
          class="combobox"
          data-season-slugs-combobox
          data-season-slugs-remove-label="${removeLabelTemplate}"
          data-auto-highlight="true"
          hidden
        >
          <input
            id="season-input"
            type="text"
            role="combobox"
            aria-expanded="false"
            aria-controls="season-listbox"
            aria-labelledby="season-label"
          />
          <div data-popover aria-hidden="true">
            <div id="season-listbox" role="listbox" aria-multiselectable="true">
              <div role="option" data-value="2024-spring">2024年春</div>
              <div role="option" data-value="2024-summer">2024年夏</div>
              <div role="option" data-value="2024-winter">2024年冬</div>
            </div>
          </div>
          <input type="hidden" />
        </div>
      </div>
    </form>
  `;

  const combobox = document.querySelector<HTMLElement>("[data-season-slugs-combobox]")!;
  const input = combobox.querySelector<HTMLInputElement>('input[role="combobox"]')!;
  const label = document.querySelector<HTMLLabelElement>("#season-label")!;
  const select = document.querySelector<HTMLSelectElement>("[data-season-slugs-select]")!;

  for (const option of select.options) {
    option.selected = selected.includes(option.value);
  }
  combobox.querySelectorAll<HTMLElement>('[role="option"]').forEach((option) => {
    if (selected.includes(option.dataset.value ?? "")) {
      option.setAttribute("aria-selected", "true");
    }
  });
  if (basecoatInitialized) {
    combobox.dataset.comboboxInitialized = "true";
  }

  return { combobox, input, label, select };
}

function selectedSlugs(select: HTMLSelectElement): string[] {
  return Array.from(select.selectedOptions, (option) => option.value);
}

function dispatchComboboxChange(combobox: HTMLElement, value: unknown): void {
  combobox.dispatchEvent(new CustomEvent("change", { detail: { value }, bubbles: true }));
}

describe("initializeSeasonFilter", () => {
  beforeEach(() => {
    document.querySelectorAll<BasecoatComboboxElement>("[data-basecoat-component]").forEach((combobox) => {
      combobox._destroy?.();
    });
    basecoatRuntime().stop();
    document.body.innerHTML = "";
  });

  it("初期化済みのcomboboxを表示し、name付きのネイティブselectを送信に使う", () => {
    const { combobox, input, label, select } = setupDom(["2024-spring"]);
    initializeSeasonFilter();

    expect(combobox.hidden).toBe(false);
    expect(select.hidden).toBe(true);
    expect(select.name).toBe("season_slugs");
    expect(select.multiple).toBe(true);
    expect(label.htmlFor).toBe(input.id);
    expect(selectedSlugs(select)).toEqual(["2024-spring"]);
  });

  it("comboboxの複数選択をselectの複数のseason_slugs値に反映する", () => {
    const { combobox, select } = setupDom();
    initializeSeasonFilter();

    dispatchComboboxChange(combobox, ["2024-spring", "2024-summer"]);

    expect(selectedSlugs(select)).toEqual(["2024-spring", "2024-summer"]);
  });

  it("comboboxの選択が空になるとselectの選択を解除する", () => {
    const { combobox, select } = setupDom(["2024-spring"]);
    initializeSeasonFilter();

    dispatchComboboxChange(combobox, []);

    expect(selectedSlugs(select)).toEqual([]);
  });

  it("選択が変更されるまでサーバー描画時のselectの初期選択を維持する", () => {
    const { select } = setupDom(["2024-spring", "2024-summer"]);
    initializeSeasonFilter();

    expect(selectedSlugs(select)).toEqual(["2024-spring", "2024-summer"]);
  });

  it("内部の入力欄から伝播するネイティブのchangeイベントを無視する", () => {
    const { input, select } = setupDom(["2024-spring"]);
    initializeSeasonFilter();

    input.dispatchEvent(new Event("change", { bubbles: true }));

    expect(selectedSlugs(select)).toEqual(["2024-spring"]);
  });

  it("Basecoatが初期化されるまでネイティブselectを表示する", () => {
    const { combobox, label, select } = setupDom(["2024-spring"], false);
    initializeSeasonFilter();

    expect(combobox.hidden).toBe(true);
    expect(select.hidden).toBe(false);
    expect(label.htmlFor).toBe(select.id);
    expect(select.name).toBe("season_slugs");
    expect(select.multiple).toBe(true);
    expect(selectedSlugs(select)).toEqual(["2024-spring"]);

    combobox.dataset.comboboxInitialized = "true";
    combobox.dispatchEvent(new CustomEvent("basecoat:initialized"));

    expect(combobox.hidden).toBe(false);
    expect(select.hidden).toBe(true);
  });

  it("閉じたBasecoatのcomboboxを再度開くときに入力した検索文字の先頭を保持する", async () => {
    const { combobox, input } = setupDom([], false);
    initializeSeasonFilter();
    basecoatRuntime().init("combobox");

    input.focus();
    input.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    expect(input.getAttribute("aria-expanded")).toBe("false");

    input.value = "春";
    input.dispatchEvent(new Event("input", { bubbles: true }));
    await new Promise<void>((resolve) => setTimeout(resolve, 0));

    expect(input.value).toBe("春");
    expect(input.getAttribute("aria-expanded")).toBe("true");
    expect(combobox.querySelector('[data-value="2024-spring"]')?.getAttribute("aria-hidden")).toBe("false");
    expect(combobox.querySelector('[data-value="2024-summer"]')?.getAttribute("aria-hidden")).toBe("true");
    expect(combobox.querySelector('[data-value="2024-winter"]')?.getAttribute("aria-hidden")).toBe("true");
  });

  it("初期選択から描画したBasecoatのチップの選択解除ラベルをローカライズする", () => {
    const { combobox } = setupDom(["2024-spring"], false, "{label}の選択を解除");
    initializeSeasonFilter();
    basecoatRuntime().init("combobox");

    expect(combobox.querySelector(".combobox-chip-remove")?.getAttribute("aria-label")).toBe("2024年春の選択を解除");
  });

  it("Basecoatで選択を変更したあとに選択解除ラベルをローカライズする", () => {
    const { combobox, input, select } = setupDom([], false, "{label}の選択を解除");
    initializeSeasonFilter();
    basecoatRuntime().init("combobox");

    input.focus();
    combobox
      .querySelector<HTMLElement>('[data-value="2024-summer"]')
      ?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    expect(selectedSlugs(select)).toEqual(["2024-summer"]);
    expect(combobox.querySelector(".combobox-chip-remove")?.getAttribute("aria-label")).toBe("2024年夏の選択を解除");
  });

  it("再度開く際にBasecoatが再描画しても選択解除ラベルのローカライズを維持する", async () => {
    const { combobox, input } = setupDom(["2024-spring"], false, "{label}の選択を解除");
    initializeSeasonFilter();
    basecoatRuntime().init("combobox");

    input.focus();
    input.dispatchEvent(new KeyboardEvent("keydown", { key: "Escape", bubbles: true }));
    input.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await Promise.resolve();

    expect(combobox.querySelector(".combobox-chip-remove")?.getAttribute("aria-label")).toBe("2024年春の選択を解除");
  });
});
