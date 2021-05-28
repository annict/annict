# frozen_string_literal: true

# Be sure to restart your server when you modify this file.

# Enable origin-checking CSRF mitigation.
# ローカルでngrok経由で開発サーバにアクセスするとき InvalidAuthenticityToken エラーを回避するため、本番環境でのみ有効にする
Rails.application.config.action_controller.forgery_protection_origin_check = Rails.env.production?
