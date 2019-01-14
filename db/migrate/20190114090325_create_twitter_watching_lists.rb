# frozen_string_literal: true

class CreateTwitterWatchingLists < ActiveRecord::Migration[5.2]
  def change
    create_table :twitter_watching_lists do |t|
      t.string :username, null: false
      t.string :name, null: false
      t.string :since_id
      t.string :discord_webhook_url, null: false
      t.timestamps null: false
    end
  end
end
