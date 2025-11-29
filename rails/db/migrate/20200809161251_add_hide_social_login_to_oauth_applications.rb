# frozen_string_literal: true

class AddHideSocialLoginToOauthApplications < ActiveRecord::Migration[6.0]
  def change
    add_column :oauth_applications, :hide_social_login, :boolean, null: false, default: false
  end
end
