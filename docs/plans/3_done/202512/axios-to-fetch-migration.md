# axios から Fetch API への移行 設計書

## 概要

Rails版フロントエンドで使用しているHTTPクライアント `axios` を削除し、ネイティブの Fetch API を使用するように移行します。

**目的**:

- 外部依存ライブラリを削減し、バンドルサイズを軽量化する
- ネイティブAPIを使用することで、ブラウザの最適化を活用する
- 既存の `fetcher.ts` ユーティリティを活用し、一貫したHTTPリクエスト処理を実現する

**背景**:

- 既に `app/javascript/utils/fetcher.ts` にFetch APIベースのラッパーが存在している
- axiosは5つのStimulusコントローラーと `application.ts` で使用されているが、使用パターンはシンプルなPOSTリクエストのみ
- Fetch APIは現代のすべてのブラウザでサポートされており、axiosの代替として十分に機能する

## 要件

### 機能要件

- 現在axiosを使用しているすべてのHTTPリクエストが、Fetch APIベースの `fetcher` ユーティリティを使用するように変更される
- CSRFトークンが自動的にリクエストヘッダーに含まれる動作が維持される
- エラーハンドリング（catchブロック）が正常に動作する
- `package.json` から `axios` 依存関係が削除される
- 既存の機能（いいね、エピソードスキップ、視聴記録など）が正常に動作する

## 設計

### 現状分析

**axiosを使用しているファイル:**

| ファイル | 使用パターン | 移行方針 |
|---------|-------------|---------|
| `application.ts` | CSRFトークンのデフォルトヘッダー設定 | 削除（fetcherが自動設定するため不要） |
| `like-button-controller.ts` | `axios.post()` | `fetcher.post()` に置換 |
| `skip-episode-button-controller.ts` | `axios.post()` | `fetcher.post()` に置換 |
| `watch-episode-button-controller.ts` | `axios.post()` | `fetcher.post()` に置換 |
| `bulk-watch-episodes-button-controller.ts` | `axios.post()` | `fetcher.post()` に置換 |
| `program-select-radio-controller.ts` | importのみ（未使用） | import文を削除 |

**既存の fetcher.ts の機能:**

- `fetcher.get(url, options)`: GETリクエスト
- `fetcher.post(url, data, options)`: POSTリクエスト
- `fetcher.delete(url, data, options)`: DELETEリクエスト
- CSRFトークンを自動的にヘッダーに付与
- エラー時にカスタムエラー（ResponseError）をthrow

### 移行方針

1. 各コントローラーの `axios` importを削除し、`fetcher` をimport
2. `axios.post()` を `fetcher.post()` に置換
3. `application.ts` からaxios関連のコードを削除
4. `package.json` から `axios` を削除
5. `yarn.lock` を更新

### コード変更例

**Before (axios使用):**
```typescript
import axios from "axios";

axios
  .post("/api/internal/likes", {
    recipient_type: this.resourceName,
    recipient_id: this.resourceId,
  })
  .then(() => {
    // 成功時の処理
  })
  .catch(() => {
    // エラー時の処理
  });
```

**After (fetcher使用):**
```typescript
import fetcher from "../utils/fetcher";

fetcher
  .post("/api/internal/likes", {
    recipient_type: this.resourceName,
    recipient_id: this.resourceId,
  })
  .then(() => {
    // 成功時の処理
  })
  .catch(() => {
    // エラー時の処理
  });
```

## タスクリスト

### フェーズ 1: コントローラーの移行

- [x] **1-1**: like-button-controller.ts の移行

  - axiosのimportを削除し、fetcherをimport
  - `axios.post()` を `fetcher.post()` に置換（2箇所）
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 10 行（実装 10 行 + テスト 0 行）

- [x] **1-2**: skip-episode-button-controller.ts の移行

  - axiosのimportを削除し、fetcherをimport
  - `axios.post()` を `fetcher.post()` に置換
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 6 行（実装 6 行 + テスト 0 行）

- [x] **1-3**: watch-episode-button-controller.ts の移行

  - axiosのimportを削除し、fetcherをimport
  - `axios.post()` を `fetcher.post()` に置換
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 6 行（実装 6 行 + テスト 0 行）

- [x] **1-4**: bulk-watch-episodes-button-controller.ts の移行

  - axiosのimportを削除し、fetcherをimport
  - `axios.post()` を `fetcher.post()` に置換
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 6 行（実装 6 行 + テスト 0 行）

- [x] **1-5**: program-select-radio-controller.ts のクリーンアップ

  - 未使用のaxios importを削除
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 2 行（実装 2 行 + テスト 0 行）

### フェーズ 2: application.ts のクリーンアップと依存関係の削除

- [x] **2-1**: application.ts からaxios関連コードを削除

  - `import axios from "axios";` を削除
  - `axios.defaults.headers.common["X-CSRF-Token"]` の設定コードを削除
  - **想定ファイル数**: 約 1 ファイル（実装 1 + テスト 0）
  - **想定行数**: 約 6 行（実装 6 行 + テスト 0 行）

- [x] **2-2**: package.json から axios 依存関係を削除

  - `"axios": "^1.1.2"` を削除
  - `yarn install` を実行して yarn.lock を更新
  - **想定ファイル数**: 約 2 ファイル（実装 2 + テスト 0）
  - **想定行数**: 約 2 行（実装 2 行 + テスト 0 行）

### フェーズ 3: 動作確認

- [ ] **3-1**: 手動テストによる動作確認

  - いいねボタンの動作確認（いいね追加/解除）
  - エピソードスキップボタンの動作確認
  - エピソード視聴記録ボタンの動作確認
  - 一括視聴記録ボタンの動作確認
  - 番組選択ラジオボタンの動作確認
  - **想定ファイル数**: 約 0 ファイル（実装 0 + テスト 0）
  - **想定行数**: 約 0 行（実装 0 行 + テスト 0 行）

### 実装しない機能（スコープ外）

以下の機能は今回の実装では**実装しません**：

- **fetcher.ts への PATCH メソッド追加**: 現在のaxios使用箇所ではPATCHメソッドは使用されていないため、今回の移行では不要
- **fetcher.ts のテスト追加**: 既存のfetcher.tsにはテストがないが、今回の移行のスコープ外
- **jQuery依存の削除**: 一部コントローラーでjQueryを使用しているが、これは別のリファクタリング対象
- **エラーハンドリングの改善**: 現在のエラーハンドリングパターンを維持し、改善は別途検討

## 参考資料

- [Fetch API - MDN Web Docs](https://developer.mozilla.org/ja/docs/Web/API/Fetch_API)
- [既存の fetcher.ts 実装](../../rails/app/javascript/utils/fetcher.ts)
