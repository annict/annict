# `app/forms/`

- フォームオブジェクトが管理されているディレクトリです
- 下記の実装方針に従っていないコードは `deprecated/` ディレクトリに格納されています

## 実装方針

- `ApplicationForm` を継承すること
  - [ActiveModel::Model](https://api.rubyonrails.org/classes/ActiveModel/Model.html) を include しています
