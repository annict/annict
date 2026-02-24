# デザインの調整 設計書

## 概要

実装済みページのデザインを統一し、ユーザーエクスペリエンス（UX）を向上させます。現在、基本的な機能は実装されていますが、デザインシステムが未整備で、ページ間のデザインの統一感が不足しています。

この設計書では、Tailwind CSS v4 と Basecoat を活用したデザインシステムの整備、共通コンポーネントの実装、レスポンシブデザインの強化、アクセシビリティの確保を行います。

**目的**:

- デザインの統一感を出し、プロフェッショナルな印象を与える
- ユーザビリティとアクセシビリティを向上させる
- 保守性の高いデザインシステムを構築する
- 開発効率を向上させる（共通コンポーネントの再利用）

**背景**:

- 現在、各ページが独立してスタイリングされており、統一感が不足
- カラーパレット、タイポグラフィ、スペーシングなどが体系化されていない
- 共通 UI コンポーネント（ボタン、カード、フォーム要素など）が定義されていない
- レスポンシブデザインは基本的には対応しているが、より洗練される必要がある
- アクセシビリティ（ARIA 属性、キーボード操作など）の配慮が不足

## 要件

### 機能要件

- デザインシステムとして、色、タイポグラフィ、スペーシング、シャドウなどのデザイントークンを定義する
- 共通 UI コンポーネント（ボタン、カード、フォーム要素、アラート、モーダルなど）のスタイルを定義する
- 既存ページ（人気アニメページ、認証ページなど）にデザインシステムを適用する
- レスポンシブデザインを強化し、モバイル、タブレット、デスクトップで最適な表示を実現する
- インタラクション（ホバー、フォーカス、アクティブ状態など）を洗練させる

### 非機能要件

#### ユーザビリティ（UX）

- 直感的で使いやすい UI を提供する
- 一貫性のあるデザインで学習コストを下げる
- 適切なフィードバック（ホバー、フォーカス、ローディング状態など）を提供する
- モバイルファーストのアプローチで、小さい画面でも使いやすい

#### アクセシビリティ

- WCAG 2.1 レベル AA に準拠する
- キーボード操作のみで全機能を利用可能にする
- スクリーンリーダーに対応する（適切な ARIA 属性の使用）
- 十分なカラーコントラスト（4.5:1 以上）を確保する
- フォーカスインジケーターを視認しやすくする

#### パフォーマンス

- CSS のファイルサイズを最小限に抑える（Tailwind CSS の Purge 機能活用）
- 不要なアニメーションやトランジションを避ける
- 画像の最適化（WebP、遅延読み込み）を維持する

#### 保守性

- Tailwind CSS のユーティリティクラスを活用し、カスタム CSS を最小限にする
- 共通コンポーネントを再利用可能な形で実装する
- デザイントークンの変更が全体に反映されやすい構造にする
- コメントやドキュメントで設計意図を明確にする

## 設計

### 技術スタック

- **Tailwind CSS v4**: ユーティリティファースト CSS フレームワーク
- **Basecoat**: Tailwind CSS ベースの UI コンポーネントライブラリ
- **Go html/template**: サーバーサイドテンプレートエンジン

### デザイントークン

Tailwind CSS v4 のデフォルトトークンをそのまま使用します。カスタムトークンは定義せず、標準のユーティリティクラスを活用することで、保守性とシンプルさを保ちます：

- **カラーパレット**: Tailwind CSS の標準カラー（gray、blue、green、red、yellow など）
- **タイポグラフィ**: Tailwind CSS の標準フォントサイズ（text-sm、text-base、text-lg など）
- **スペーシング**: Tailwind CSS の標準スペーシング（p-4、m-8 など）
- **ブレークポイント**: Tailwind CSS の標準ブレークポイント（sm: 640px、md: 768px、lg: 1024px、xl: 1280px など）
- **シャドウ**: Tailwind CSS の標準シャドウ（shadow-sm、shadow、shadow-md など）
- **角丸**: Tailwind CSS の標準ボーダーラディウス（rounded、rounded-md、rounded-lg など）

### ダークモード対応

Rails 版の Annict ではダークモード対応がされているため、Go 版でも最初からダークモード対応を行います。

**実装方針**:

- **検出方法**: `prefers-color-scheme`メディアクエリを使用してシステムの設定を自動検出
- **Tailwind CSS の dark バリアント**: すべてのコンポーネントで`dark:`プレフィックスを使用
- **カラースキーム**:
  - ライトモード: 白ベース（bg-white、bg-gray-50 など）
  - ダークモード: 暗色ベース（bg-gray-900、bg-gray-800 など）
- **テキストカラー**:
  - ライトモード: text-gray-900、text-gray-600 など
  - ダークモード: text-gray-100、text-gray-300 など
- **ボーダーとシャドウ**:
  - ライトモード: border-gray-200、shadow-sm など
  - ダークモード: border-gray-700、dark:shadow-gray-900/50 など

**実装例**:

```html
<!-- ライトモード: bg-white、ダークモード: bg-gray-900 -->
<div class="bg-white dark:bg-gray-900">
  <!-- ライトモード: text-gray-900、ダークモード: text-gray-100 -->
  <h1 class="text-gray-900 dark:text-gray-100">タイトル</h1>
</div>
```

**将来の拡張性**:

- 現在は`prefers-color-scheme`による自動検出のみ
- 将来的にユーザー設定（手動切り替え）を追加する可能性を考慮した設計

### 共通スタイル定義

Basecoat CSS と Tailwind CSS の `@layer components` を活用して、再利用可能なユーティリティクラスを定義します。

**実装方針**:

- `web/style.css` で `@layer components` を使用してスタイルクラスを定義
- Tailwind CSS の `@apply` ディレクティブで既存のユーティリティクラスを組み合わせる
- 必要に応じて `{{define}}` でテンプレートパーツを作成（繰り返しが多い場合のみ）

**定義するクラスカテゴリ**:

1. **リンク**: `.link`, `.link-primary`, `.link-muted`
2. **ボタン**: `.btn-primary`, `.btn-secondary`, `.btn-outline`, `.btn-text`
3. **フォーム要素**: `.form-input`, `.form-textarea`, `.form-select`, `.form-checkbox`, `.form-radio`
4. **カード**: `.card`, `.card-hover`, `.card-link`
5. **アラート**: `.alert-success`, `.alert-error`, `.alert-warning`, `.alert-info`
6. **レイアウト**: `.container`, `.section`

**実装例**:

```css
/* web/style.css */
@layer components {
  /* リンク */
  .link {
    @apply text-mizuho-600 hover:text-mizuho-700 underline;
    @apply dark:text-mizuho-400 dark:hover:text-mizuho-300;
    @apply transition-colors duration-150;
  }

  .link-muted {
    @apply text-stone-600 hover:text-stone-900;
    @apply dark:text-stone-400 dark:hover:text-stone-100;
    @apply transition-colors duration-150;
  }

  /* ボタン */
  .btn-primary {
    @apply px-4 py-2 bg-mizuho-600 text-white rounded-lg;
    @apply hover:bg-mizuho-700 focus:ring-2 focus:ring-mizuho-500 focus:outline-none;
    @apply dark:bg-mizuho-500 dark:hover:bg-mizuho-600;
    @apply transition-colors duration-150;
    @apply disabled:opacity-50 disabled:cursor-not-allowed;
  }

  /* フォーム要素 */
  .form-input {
    @apply w-full px-3 py-2 border border-stone-300 rounded-lg;
    @apply focus:ring-2 focus:ring-mizuho-500 focus:border-mizuho-500 focus:outline-none;
    @apply dark:bg-stone-800 dark:border-stone-600 dark:text-stone-100;
    @apply transition-colors duration-150;
  }

  /* カード */
  .card {
    @apply bg-white rounded-lg shadow-sm border border-stone-200;
    @apply dark:bg-stone-800 dark:border-stone-700;
  }

  .card-hover {
    @apply card transition-all duration-200;
    @apply hover:shadow-md hover:-translate-y-0.5;
  }
}
```

**使用例**:

```html
<!-- リンク -->
<a href="/works" class="link">作品一覧</a>

<!-- ボタン -->
<button type="submit" class="btn-primary">送信</button>

<!-- フォーム -->
<input type="text" class="form-input" placeholder="メールアドレス" />

<!-- カード -->
<div class="card-hover p-4">
  <h3>タイトル</h3>
</div>
```

### レイアウト

既存の 3 カラムレイアウト（ヘッダー、メイン、フッター）を維持しつつ、以下を改善：

- **ヘッダー**: ロゴ、ナビゲーション、ユーザーメニュー
- **メインコンテンツ**: 最大幅を設定し、読みやすさを向上
- **フッター**: 著作権表示、リンク集

### レスポンシブデザイン

Tailwind CSS のブレークポイントを活用：

- **モバイル（< 640px）**: 1 カラムレイアウト、スタックされたナビゲーション
- **タブレット（640px - 1024px）**: 2〜3 カラムのグリッド、折りたたみ可能なナビゲーション
- **デスクトップ（> 1024px）**: 4〜5 カラムのグリッド、完全なナビゲーション

### アクセシビリティ対策

- **セマンティック HTML**: `<header>`, `<nav>`, `<main>`, `<footer>`などの要素を使用
- **ARIA 属性**: `aria-label`, `aria-describedby`, `aria-current`などを適切に使用
- **キーボード操作**: すべてのインタラクティブ要素がキーボードで操作可能
- **フォーカスインジケーター**: `focus:ring`と`focus:outline`で視認性を確保
- **カラーコントラスト**: WCAG AA レベル（4.5:1 以上）を確保

### 実装方針

1. **段階的な改善**: 既存ページを破壊せず、段階的にデザインを改善
2. **共通化優先**: 複数ページで使用されるコンポーネントを優先的に実装
3. **テスト駆動**: デザイン変更後もテストが通ることを確認
4. **ドキュメント化**: 共通コンポーネントの使い方をコメントで明記

## タスクリスト

### フェーズ 1: 共通ユーティリティクラスの定義

`web/style.css` に `@layer components` で共通スタイルを定義します。

**注**: Basecoat が提供しているボタン、フォーム、カード、アラートのスタイルはそのまま使用します。

- [x] **リンクスタイルの定義**
  - `.link`（基本リンク）: 青色でアンダーライン付き
  - ダークモード対応（`dark:`バリアント）
  - ホバー、フォーカス状態のスタイル
  - **想定ファイル数**: 1 ファイル（web/style.css）
  - **想定行数**: 約 30 行

---

### フェーズ 2: 既存ページへのスタイル適用

Basecoat のスタイルと定義したユーティリティクラスを既存のテンプレートに適用します。

- [x] **サインインページの改善**

  - `internal/templates/sign_in.html` のスタイル更新
  - フォーム要素に Basecoat のフォームスタイルを適用
  - ボタンに Basecoat のボタンスタイルを適用
  - リンクに `.link` を適用
  - エラーメッセージに Basecoat のアラートスタイルを適用
  - レイアウトとスペーシングの調整
  - ダークモード対応
  - **想定ファイル数**: 1 ファイル
  - **想定行数**: 約 60 行（既存コードの修正）

- [x] **パスワードリセットページの改善**

  - `internal/templates/password/reset.html` のスタイル更新（リセット申請）
  - `internal/templates/password/edit.html` のスタイル更新（パスワード変更）
  - `internal/templates/password/reset_sent.html` のスタイル更新（送信完了）
  - フォーム要素、ボタン、リンク、アラートに Basecoat のスタイルを適用
  - レイアウトの統一
  - ダークモード対応
  - **想定ファイル数**: 3 ファイル
  - **想定行数**: 約 100 行（既存コードの修正）

---

### 未対応タスクの分割について

以下のタスクは詳細な設計書として分割されました：

- **ベースレイアウト、人気アニメページ、エラーページ、メールテンプレートの改善** → `.claude/designs/2_todo/ui-layout-improvements.md` を参照
- **アクセシビリティの強化（キーボード操作、ARIA 属性、セマンティック HTML、カラーコントラスト）** → `.claude/designs/2_todo/accessibility-improvements.md` を参照
- **パフォーマンス最適化とドキュメント化（CSS 最適化、アニメーション、スタイルガイド）** → `.claude/designs/2_todo/performance-optimization.md` を参照

---

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **ユーザーによるダークモード手動切り替え**: 現在は `prefers-color-scheme` による自動検出のみ。将来的にユーザー設定画面で手動切り替えを追加する可能性あり
- **複雑なアニメーション**: パフォーマンスへの影響を考慮し、シンプルなトランジションのみ実装
- **カスタムフォント**: システムフォントを使用し、読み込み時間を短縮
- **高度なインタラクション**: JavaScript 依存の複雑な UI は、htmx 導入時に検討
- **テンプレートパーツ（`{{define}}`）の積極的な利用**: 繰り返しが多い場合のみ検討し、基本は CSS のユーティリティクラスで対応

## 参考資料

- [Tailwind CSS v4 Documentation](https://tailwindcss.com/docs)
- [Tailwind CSS - Dark Mode](https://tailwindcss.com/docs/dark-mode)
- [Basecoat CSS](https://basecoat.style/)
- [WCAG 2.1 ガイドライン](https://www.w3.org/WAI/WCAG21/quickref/)
- [Material Design Guidelines](https://material.io/design)
- [Refactoring UI (書籍)](https://www.refactoringui.com/)
- [prefers-color-scheme - MDN](https://developer.mozilla.org/ja/docs/Web/CSS/@media/prefers-color-scheme)

---

## 注意事項

- デザイン変更時は、必ず既存のテストが通ることを確認してください
- 各タスクは 1 つの Pull Request で完結するように実装してください
- PR 作成時は、変更前後のスクリーンショットを添付してください
- アクセシビリティ対応は優先度が高いため、フェーズ 7 を早めに実施することを推奨します
