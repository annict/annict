# frozen_string_literal: true

class CreateCollections < ActiveRecord::Migration[5.0]
  def change
    create_table :collections do |t|
      t.integer :user_id, null: false
      t.string :title, null: false
      t.string :description
      t.string :aasm_state, null: false, default: "draft"
      t.integer :likes_count, null: false, default: 0
      t.datetime :published_at
      t.timestamps null: false
    end

    add_index :collections, :user_id
    add_foreign_key :collections, :users
  end
end
