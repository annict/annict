# Annict プロジェクト概要

## プロジェクトの目的

Annictはアニメ視聴記録サービスです。ユーザーは見たアニメを記録したり、アニメの情報を検索したりすることができます。

## 技術スタック

- **バックエンド**: Ruby 3.3.5, Ruby on Rails 7.1.0
- **フロントエンド**: TypeScript, Hotwire (Stimulus + Turbo), Tailwind CSS
- **パッケージマネージャー**: Bundler (Ruby), pnpm (JavaScript)
- **テスティングフレームワーク**: RSpec, FactoryBot
- **Linter**: Standard (Ruby), ESLint (JavaScript), ERB Lint (ERB)
- **型検査**: Sorbet (Ruby), TypeScript
- **データベース**: PostgreSQL

## 主要な依存関係

- ViewComponent (UI コンポーネント)
- Devise (認証)
- Doorkeeper (OAuth2)
- GraphQL
- Active Decorator
- Kaminari (ページネーション)
- Shrine (ファイルアップロード with AWS S3)

## 開発環境

- Ruby 3.3.5
- Node.js 22.17.0
- Yarn 1.22.22
- システム: Darwin (macOS)
- 現在のブランチ: rails-7-1 (メインブランチ: main)

## ライセンス

Apache License 2.0 (2014-2024 Annict)
