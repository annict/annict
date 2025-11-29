# frozen_string_literal: true

class CreateGumroadSubscribers < ActiveRecord::Migration[5.1]
  def change
    create_table :gumroad_subscribers do |t|
      t.string :gumroad_id, null: false
      t.string :gumroad_product_id, null: false
      t.string :gumroad_product_name, null: false
      t.string :gumroad_user_id, null: false
      t.string :gumroad_user_email, null: false
      t.string :gumroad_purchase_ids, null: false, array: true
      t.datetime :gumroad_created_at, null: false
      t.datetime :gumroad_cancelled_at
      t.datetime :gumroad_user_requested_cancellation_at
      t.datetime :gumroad_charge_occurrence_count
      t.datetime :gumroad_ended_at
      t.timestamps null: false
    end

    add_index :gumroad_subscribers, :gumroad_id, unique: true
    add_index :gumroad_subscribers, :gumroad_product_id
    add_index :gumroad_subscribers, :gumroad_user_id
  end
end
