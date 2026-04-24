---
paths:
  - "rails/**/*.{rb,erb}"
---

# アーキテクチャガイド

このドキュメントは、Rails 版 Wikino のクラス設計と依存関係のルールを説明します。

## クラス設計と依存関係

### クラス間の依存関係ルール

| クラス     | 依存可能な先                                   |
| ---------- | ---------------------------------------------- |
| Component  | Component, Form, Model                         |
| Controller | Form, Model, Record, Repository, Service, View |
| Form       | Record, Validator                              |
| Job        | Service                                        |
| Mailer     | Model, Record, Repository, View                |
| Model      | Model                                          |
| Policy     | Record                                         |
| Record     | Record                                         |
| Repository | Model, Record, Policy                          |
| Service    | Job, Mailer, Record                            |
| Validator  | Record                                         |
| View       | Component, Form, Model                         |

### Service と Job の依存関係について

Service と Job の間には相互依存が存在しますが、以下のルールで循環依存を回避します：

- **Service → Job**: `perform_later`メソッドによるキューへの追加のみ許可
- **Job → Service**: ジョブ実行時の Service 呼び出しは許可
- **重要**: Service から Job インスタンスの直接実行（`perform`メソッド）は禁止

```ruby
# ✅ 良い例：Serviceからジョブをキューに追加
class Users::CreateService
  def call
    user = UserRecord.create!(...)
    Users::SendWelcomeEmailJob.perform_later(user.id)  # キューに追加のみ
  end
end

# ❌ 悪い例：Serviceからジョブを直接実行
class Users::CreateService
  def call
    user = UserRecord.create!(...)
    Users::SendWelcomeEmailJob.new.perform(user.id)  # 直接実行は禁止
  end
end
```

### 命名規則

- Controller: `(ModelPlural)::(ActionName)Controller`
- Service: `(ModelPlural)::(Verb)Service`
- Form: `(ModelPlural)::(Noun)Form`
- Repository: `(Model)Repository`
- View: `(ModelPlural)::(ActionName)View`
- Component: `(UIComponentPlural)::(Noun)Component`

## サービスクラスのルール

### サービスクラスを使用する場合

- ✅ データベースへの永続化を伴う処理
- ✅ 複数のモデル/レコードにまたがる複雑なビジネスロジックで永続化を伴うもの
- ✅ トランザクション管理が必要な処理

### サービスクラスを使用しない場合

- ❌ データベースへの永続化を伴わない処理（URL 生成、データ変換など）
- ❌ 単一のモデル/レコードに閉じた処理（モデルやレコードのメソッドとして定義）

### トランザクション処理

**重要**: Service クラスでトランザクションを張る場合は、必ず `#with_transaction` メソッドを使用すること

```ruby
# ✅ 良い例：with_transactionを使用
module Users
  class CreateService < ApplicationService
    def call
      with_transaction do
        user = UserRecord.create!(...)
        ProfileRecord.create!(user:, ...)
      end
    end
  end
end

# ❌ 悪い例：transactionを直接使用
module Users
  class CreateService < ApplicationService
    def call
      ApplicationRecord.transaction do
        # with_transactionを使うべき
      end
    end
  end
end
```

**重要**: Controller、Job、Rake タスク内で永続化処理を書く場合は、必ず Service クラスを定義すること

## 重要な原則

- ネストしたトランザクションを避ける
- レコードのコールバックを避ける
- View/Component でのデータベースアクセスを防ぐ
- 問題が解決されるなら、レイヤーを跨いだ依存も許可
