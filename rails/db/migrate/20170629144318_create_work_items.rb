# frozen_string_literal: true

class CreateWorkItems < ActiveRecord::Migration[5.1]
  def change
    create_table :work_items do |t|
      t.integer :work_id, null: false
      t.integer :item_id, null: false
      t.integer :user_id, null: false
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :work_items, :work_id
    add_index :work_items, :item_id
    add_index :work_items, :user_id
    add_index :work_items, %i[work_id item_id], unique: true

    add_foreign_key :work_items, :works
    add_foreign_key :work_items, :items
    add_foreign_key :work_items, :users
  end
end
