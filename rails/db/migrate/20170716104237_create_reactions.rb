# frozen_string_literal: true

class CreateReactions < ActiveRecord::Migration[5.1]
  def change
    create_table :reactions do |t|
      t.integer :user_id, null: false
      t.integer :target_user_id, null: false
      t.string :kind, null: false
      t.integer :collection_item_id
      t.timestamps null: false
    end

    add_index :reactions, :user_id
    add_index :reactions, :target_user_id
    add_index :reactions, :collection_item_id

    add_foreign_key :reactions, :users
    add_foreign_key :reactions, :users, column: :target_user_id
    add_foreign_key :reactions, :collection_items
  end
end
