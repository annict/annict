# frozen_string_literal: true

class AddShareStatus < ActiveRecord::Migration[5.1]
  def change
    add_column :settings, :share_status_to_twitter, :boolean, null: false, default: false
    add_column :settings, :share_status_to_facebook, :boolean, null: false, default: false
  end
end
