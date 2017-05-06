# frozen_string_literal: true

class CreateCollectionItems < ActiveRecord::Migration[5.0]
  def change
    create_table :collection_items do |t|
      t.integer :collection_id, null: false
      t.integer :work_id, null: false
      t.string :title, null: false
      t.text :comment
      t.timestamps null: false
    end

    add_index :collection_items, :collection_id
    add_index :collection_items, :work_id
    add_index :collection_items, %i(collection_id work_id), unique: true

    add_foreign_key :collection_items, :collections
    add_foreign_key :collection_items, :works
  end
end
