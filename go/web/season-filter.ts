// Progressively enhances the durable native release-season select with Basecoat's searchable,
// chip-based multiple combobox. The named select remains in the form and is kept in sync, so
// GET submissions use repeated season_slugs parameters. The native control stays visible when
// Basecoat or this bridge does not initialize.
//
// [Ja] 常用可能なネイティブのリリース時期 select を、Basecoat の検索・チップ式の複数選択
// combobox でプログレッシブエンハンスメントする。name 付き select はフォーム内に残して
// 同期するため、GET 送信では繰り返し season_slugs パラメータを使う。Basecoat または本
// ブリッジが初期化されなければ、ネイティブコントロールを表示したままにする。

const comboboxSelector = "[data-season-slugs-combobox]";
const filterSelector = "[data-season-slugs-filter]";
const selectSelector = "[data-season-slugs-select]";
const removeLabelPlaceholder = "{label}";

interface SeasonSlugsChangeDetail {
  value?: unknown;
}

export function initializeSeasonFilter(): void {
  document.querySelectorAll<HTMLElement>(comboboxSelector).forEach((combobox) => {
    if (combobox.dataset.seasonFilterBound === "true") {
      return;
    }
    combobox.dataset.seasonFilterBound = "true";

    if (combobox.dataset.comboboxInitialized === "true") {
      enhanceSeasonFilter(combobox);
      return;
    }

    combobox.addEventListener("basecoat:initialized", () => enhanceSeasonFilter(combobox), { once: true });
  });
}

function enhanceSeasonFilter(combobox: HTMLElement): void {
  if (combobox.dataset.seasonFilterEnhanced === "true") {
    return;
  }

  const select = findSelect(combobox);
  const input = combobox.querySelector<HTMLInputElement>('input[role="combobox"]');
  if (!select || !input) {
    return;
  }

  combobox.addEventListener("change", handleChange);
  input.addEventListener("input", preserveSearchOnReopen, { capture: true });
  input.addEventListener("focus", scheduleChipRemoveLabelLocalization);
  input.addEventListener("click", scheduleChipRemoveLabelLocalization);
  input.addEventListener("input", scheduleChipRemoveLabelLocalization);
  input.addEventListener("keydown", scheduleChipRemoveLabelLocalization);

  const labelID = input.getAttribute("aria-labelledby");
  const label = labelID ? document.getElementById(labelID) : null;
  if (label instanceof HTMLLabelElement) {
    label.htmlFor = input.id;
  }

  localizeChipRemoveLabels(combobox);

  combobox.dataset.seasonFilterEnhanced = "true";
  combobox.hidden = false;
  select.hidden = true;
}

function handleChange(event: Event): void {
  // Basecoat dispatches its selection-change CustomEvent on the combobox root, so ignore
  // native change events bubbling up from the inner text input.
  //
  // [Ja] Basecoat は選択変更の CustomEvent を combobox の root で発火するため、内側の
  // テキスト input から浮上するネイティブ change は無視する。
  if (event.target !== event.currentTarget) {
    return;
  }

  const combobox = event.currentTarget as HTMLElement;
  const detail = (event as CustomEvent<SeasonSlugsChangeDetail>).detail;
  syncSelect(combobox, toSlugs(detail?.value));
  localizeChipRemoveLabels(combobox);
}

// Preserve the first character typed when a closed Basecoat combobox reopens.
// Basecoat refreshes the multiple selection while opening and clears the input.
//
// [Ja] 閉じた Basecoat combobox を再度開くときに入力した最初の文字を保持する。
// Basecoat は開く際に複数選択を再描画し、入力欄を空にする。
function preserveSearchOnReopen(event: Event): void {
  const input = event.currentTarget;
  if (!(input instanceof HTMLInputElement) || input.getAttribute("aria-expanded") !== "false" || input.value === "") {
    return;
  }

  const search = input.value;
  setTimeout(() => {
    if (input.value !== "") {
      return;
    }

    input.value = search;
    input.dispatchEvent(new Event("input", { bubbles: true }));
  }, 0);
}

// Localize Basecoat-generated chip remove labels after initial rendering and changes.
//
// [Ja] Basecoat が生成するチップの選択解除ラベルを初期描画後と変更後にローカライズする。
function localizeChipRemoveLabels(combobox: HTMLElement): void {
  const labelTemplate = combobox.dataset.seasonSlugsRemoveLabel;
  if (!labelTemplate?.includes(removeLabelPlaceholder)) {
    return;
  }

  combobox.querySelectorAll<HTMLButtonElement>(".combobox-chip-remove").forEach((button) => {
    const label = button.previousElementSibling?.textContent?.trim();
    if (!label) {
      return;
    }

    button.setAttribute("aria-label", labelTemplate.replace(removeLabelPlaceholder, label));
  });
}

// Reapply localized labels after Basecoat reopens and refreshes the combobox.
//
// [Ja] Basecoat が combobox を再度開いて再描画したあと、ローカライズ済みラベルを再適用する。
function scheduleChipRemoveLabelLocalization(event: Event): void {
  const input = event.currentTarget;
  if (!(input instanceof HTMLInputElement)) {
    return;
  }

  const combobox = input.closest<HTMLElement>(comboboxSelector);
  if (!combobox) {
    return;
  }

  queueMicrotask(() => localizeChipRemoveLabels(combobox));
}

function toSlugs(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map(String);
  }
  if (typeof value === "string" && value !== "") {
    return [value];
  }
  return [];
}

function syncSelect(combobox: HTMLElement, slugs: string[]): void {
  const select = findSelect(combobox);
  if (!select) {
    return;
  }

  const selectedSlugs = new Set(slugs);
  for (const option of select.options) {
    option.selected = selectedSlugs.has(option.value);
  }
}

function findSelect(combobox: HTMLElement): HTMLSelectElement | null {
  return combobox.closest(filterSelector)?.querySelector<HTMLSelectElement>(selectSelector) ?? null;
}
