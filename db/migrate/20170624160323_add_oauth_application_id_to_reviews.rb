# frozen_string_literal: true

class AddOauthApplicationIdToReviews < ActiveRecord::Migration[5.1]
  def change
    add_column :reviews, :oauth_application_id, :integer
    add_index :reviews, :oauth_application_id
    add_foreign_key :reviews, :oauth_applications
  end
end
