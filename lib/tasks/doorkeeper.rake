# frozen_string_literal: true

namespace :doorkeeper do
  # 非公開になっている or ユーザが退会した (owner_idが空) のOauth::Application から
  # 生成されたアクセストークンを無効にする
  task revoke_access_tokens: :environment do
    Oauth::Application.unavailable.find_each do |app|
      app.access_tokens.where(revoked_at: nil).each(&:revoke)
    end
  end
end
