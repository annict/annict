---
name: htmx4
description: htmx 4 (alpha) を使った機能の実装・リファクタリングを支援する。HTML フラグメントの返却パターンによるサーバードリブン UI の実装、templ テンプレートでの htmx 属性の使用、OOB スワップによる複数要素の同時更新を正確に行える。
argument-hint: "[実装対象の機能説明]"
---

# htmx 4 スキル

htmx 4 を使った機能の実装・リファクタリングを行います。

**実装対象**: $ARGUMENTS

## このスキルの使い方

1. 下記の「主要な属性リファレンス」と「Go 統合パターン」を参照する
2. 不明な点がある場合はソースコード https://cdn.jsdelivr.net/npm/htmx.org@4.0.0-beta2/dist/htmx.js を参照する

## ファイル構成

```
skills/htmx4/
└── SKILL.md                           # このファイル
```

## 主要な属性リファレンス

### リクエスト発行

| 属性        | 説明                          |
| ----------- | ----------------------------- |
| `hx-get`    | 指定 URL に GET リクエスト    |
| `hx-post`   | 指定 URL に POST リクエスト   |
| `hx-patch`  | 指定 URL に PATCH リクエスト  |
| `hx-delete` | 指定 URL に DELETE リクエスト |

### レスポンスの制御

| 属性          | デフォルト  | 説明                         |
| ------------- | ----------- | ---------------------------- |
| `hx-target`   | `this`      | レスポンスを挿入する要素     |
| `hx-swap`     | `innerHTML` | レスポンスの挿入方法         |
| `hx-swap-oob` | -           | ターゲット外の要素を同時更新 |

### hx-swap の値

| 値            | 説明                                 |
| ------------- | ------------------------------------ |
| `innerHTML`   | 要素の内部 HTML を置換（デフォルト） |
| `outerHTML`   | 要素全体を置換                       |
| `beforeend`   | 要素の末尾に追加（`append`）         |
| `afterbegin`  | 要素の先頭に追加（`prepend`）        |
| `beforebegin` | 要素の前に挿入（`before`）           |
| `afterend`    | 要素の後に挿入（`after`）            |
| `innerMorph`  | 内部 HTML を morph で置換            |
| `outerMorph`  | 要素全体を morph で置換              |
| `delete`      | 要素を削除（レスポンス無視）         |
| `none`        | 挿入しない（OOB スワップは動作する） |

### トリガー制御

```html
<!-- デフォルト: input/textarea/select は change、form は submit、その他は click -->
<button hx-get="/items">更新</button>

<!-- カスタムイベントでトリガー -->
<div hx-get="/draft" hx-trigger="draft-autosaved from:window"></div>

<!-- ポーリング -->
<div hx-get="/updates" hx-trigger="every 2s"></div>

<!-- 修飾子 -->
<input hx-get="/search" hx-trigger="input changed delay:500ms" />
```

主要な修飾子: `once`, `changed`, `delay:<time>`, `throttle:<time>`, `from:<selector>`, `consume`

### フォーム送信中の要素無効化

```html
<!-- htmx 4 では hx-disabled-elt が hx-disable にリネームされた -->
<button hx-post="/submit" hx-disable="this">送信</button>

<!-- フォーム内の複数要素を無効化 -->
<form hx-post="/submit" hx-disable="find input, find button">
  <input type="text" name="title" />
  <button type="submit">送信</button>
</form>
```

### OOB（Out-Of-Band）スワップ

レスポンスに `hx-swap-oob` 属性を持つ要素を含めると、メインターゲット以外の要素も同時に更新できる。

```html
<!-- レスポンス HTML -->
<!-- メインターゲットに挿入される内容 -->
<span id="saved-at">保存済み: 12:34</span>

<!-- OOB スワップ: id が一致する要素を置換 -->
<div id="link-list" hx-swap-oob="innerHTML">...新しいリンク一覧...</div>
<div id="backlink-list" hx-swap-oob="innerHTML">...新しいバックリンク一覧...</div>
```

## Go 統合パターン

htmx は「サーバーから HTML フラグメントを返す」だけで動作する。専用の Go SDK は不要。

### 基本パターン: HTML フラグメントを返す

```go
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // データを取得
    items, err := h.getItemsUsecase.Execute(ctx, input)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // templ コンポーネントをレンダリングして返す（通常の HTML レスポンス）
    component := components.ItemList(viewmodel.NewItems(items))
    component.Render(ctx, w)
}
```

### OOB スワップパターン: 複数要素の同時更新

```go
func (h *Handler) Show(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    data, err := h.getDataUsecase.Execute(ctx, input)
    if err != nil {
        http.Error(w, "Internal Server Error", http.StatusInternalServerError)
        return
    }

    // メインターゲット + OOB 要素を含むコンポーネントをレンダリング
    component := components.DraftPageResponse(data)
    component.Render(ctx, w)
}
```

```templ
// OOB スワップ用のレスポンステンプレート
templ DraftPageResponse(data DraftPageData) {
    // メインターゲット（hx-target で指定された要素に挿入される）
    <span id="page-draft-saved-at">{ data.SavedAt }</span>

    // OOB スワップ（id が一致する要素を自動的に更新）
    <div id="page-link-list" hx-swap-oob="innerHTML">
        @LinkList(data.Links)
    </div>
    <div id="page-backlink-list" hx-swap-oob="innerHTML">
        @BacklinkList(data.Backlinks)
    </div>
}
```

### ページネーション「もっと読み込む」パターン

```templ
templ LoadMoreButton(nextPage int, url string) {
    <div id="load-more">
        <button hx-get={ fmt.Sprintf("%s?page=%d", url, nextPage) }
                hx-target="#item-list"
                hx-swap="beforeend"
                hx-select="oob">
            もっと読み込む
        </button>
    </div>
}
```

レスポンスにはアイテム一覧と、更新された「もっと読み込む」ボタン（OOB）を含める。

### フォーム送信（hx-post）パターン

```templ
templ Form(data FormData) {
    <form hx-post={ data.Action } hx-disable="find button[type='submit']">
        <input type="hidden" name="csrf_token" value={ data.CSRFToken }/>
        <input type="text" name="title"/>
        <button type="submit">送信</button>
    </form>
}
```

```go
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()

    // バリデーション失敗時: フォームを再レンダリング（htmx 4 では全ステータスコードがスワップされる）
    if result.FormErrors.HasErrors() {
        w.WriteHeader(http.StatusUnprocessableEntity)
        component := pages.New(formDataWithErrors)
        component.Render(ctx, w)
        return
    }

    // 成功時: HX-Redirect ヘッダーでリダイレクト
    w.Header().Set("HX-Redirect", redirectPath)
    w.WriteHeader(http.StatusNoContent)
}
```

### htmx リクエストの判別

```go
// htmx からのリクエストかどうかを判別
func isHTMXRequest(r *http.Request) bool {
    return r.Header.Get("HX-Request") == "true"
}
```

## CSRF 対策

htmx はフォーム送信と同様に `<input type="hidden" name="csrf_token">` でトークンを送信するため、既存の CSRF ミドルウェアがそのまま機能する。Datastar のようにシグナルから読み取る特別な処理は不要。

```templ
<form hx-post="/items">
    <input type="hidden" name="csrf_token" value={ data.CSRFToken }/>
    <!-- フォームフィールド -->
</form>
```

## htmx 4 の重要な変更点（htmx 2 との違い）

| 項目                 | htmx 2                   | htmx 4                              |
| -------------------- | ------------------------ | ----------------------------------- |
| エラーレスポンス     | 4xx/5xx はスワップしない | 204/304 以外すべてスワップする      |
| 属性の継承           | 暗黙的（自動継承）       | 明示的（`:inherited` 修飾子が必要） |
| `hx-disabled-elt`    | 使用可能                 | `hx-disable` にリネーム             |
| OOB スワップの順序   | OOB が先                 | メインが先、OOB が後                |
| GET のフォームデータ | 含む                     | 含まない                            |
| `hx-ext`             | 属性で指定               | `<script>` で直接読み込み           |

## レスポンスヘッダー

| ヘッダー         | 説明                                        |
| ---------------- | ------------------------------------------- |
| `HX-Redirect`    | クライアント側でリダイレクト                |
| `HX-Refresh`     | ページをリフレッシュ（`"true"` を設定）     |
| `HX-Push-Url`    | ブラウザの URL バーを更新                   |
| `HX-Replace-Url` | ブラウザの URL バーを置換（履歴に残さない） |
| `HX-Trigger`     | クライアント側でイベントを発火              |

## 実装チェックリスト

### 1. HTML テンプレート（templ）の準備

- [ ] 更新対象の要素に一意の `id` 属性を付与する
- [ ] `hx-get` / `hx-post` で HTML フラグメントを返すエンドポイントを指定する
- [ ] `hx-target` でレスポンスの挿入先を指定する
- [ ] `hx-swap` でレスポンスの挿入方法を指定する
- [ ] フォーム送信では `hx-disable` で二重送信を防止する

### 2. Go ハンドラーの実装

- [ ] templ コンポーネントをレンダリングして HTML フラグメントを返す
- [ ] 複数要素の更新が必要な場合は OOB スワップ用の要素をレスポンスに含める
- [ ] フォーム送信成功時は `HX-Redirect` ヘッダーでリダイレクトする
- [ ] バリデーションエラー時は 422 ステータスでフォームを再レンダリングする

### 3. CSRF 対策の確認

- [ ] フォームに `<input type="hidden" name="csrf_token">` を含める
- [ ] 既存の CSRF ミドルウェアが POST/PATCH/DELETE で自動検証する
