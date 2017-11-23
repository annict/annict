# frozen_string_literal: true

class AddGumroadSubscriberIdToUsers < ActiveRecord::Migration[5.1]
  def change
    add_column :users, :gumroad_subscriber_id, :integer
    add_index :users, :gumroad_subscriber_id
    add_foreign_key :users, :gumroad_subscribers
  end
end
