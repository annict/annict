# frozen_string_literal: true

class AddOfficialToOauthApplications < ActiveRecord::Migration[5.2]
  def change
    add_column :oauth_applications, :official, :boolean, null: false, default: false
  end
end
