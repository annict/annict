# frozen_string_literal: true

class CreateWorkTaggings < ActiveRecord::Migration[5.1]
  def change
    create_table :work_taggings do |t|
      t.integer :user_id, null: false
      t.integer :work_id, null: false
      t.integer :work_tag_id, null: false
      t.timestamps null: false
    end

    add_index :work_taggings, :user_id
    add_index :work_taggings, :work_id
    add_index :work_taggings, :work_tag_id
    add_index :work_taggings, %i[user_id work_id work_tag_id], unique: true
    add_foreign_key :work_taggings, :users
    add_foreign_key :work_taggings, :works
    add_foreign_key :work_taggings, :work_tags
  end
end
