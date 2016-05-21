# frozen_string_literal: true

class AddOauthApplicationIdToCheckinsAndStatuses < ActiveRecord::Migration
  def change
    add_column :checkins, :oauth_application_id, :integer
    add_index :checkins, :oauth_application_id
    add_foreign_key :checkins, :oauth_applications
    add_column :statuses, :oauth_application_id, :integer
    add_index :statuses, :oauth_application_id
    add_foreign_key :statuses, :oauth_applications
  end
end
