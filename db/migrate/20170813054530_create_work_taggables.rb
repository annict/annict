# frozen_string_literal: true

class CreateWorkTaggables < ActiveRecord::Migration[5.1]
  def change
    create_table :work_taggables do |t|
      t.integer :user_id, null: false
      t.integer :work_tag_id, null: false
      t.string :description
      t.timestamps null: false
    end

    add_index :work_taggables, :user_id
    add_index :work_taggables, :work_tag_id
    add_index :work_taggables, %i[user_id work_tag_id], unique: true

    add_foreign_key :work_taggables, :users
    add_foreign_key :work_taggables, :work_tags
  end
end
