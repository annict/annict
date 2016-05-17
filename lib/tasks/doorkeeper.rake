# frozen_string_literal: true

namespace :doorkeeper do
  # 非公開になっている or ユーザが退会した (owner_idが空) のDoorkeeper::Application から
  # 生成されたアクセストークンを無効にする
  task revoke_access_tokens: :environment do
    Doorkeeper::Application.unavailable.find_each do |app|
      app.access_tokens.each(&:revoke)
    end
  end
end
