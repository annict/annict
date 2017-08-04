# frozen_string_literal: true

class CreateUserWorkTags < ActiveRecord::Migration[5.1]
  def change
    create_table :user_work_tags do |t|
      t.integer :user_id, null: false
      t.integer :work_id, null: false
      t.integer :work_tag_id, null: false
      t.timestamps null: false
    end

    add_index :user_work_tags, :user_id
    add_index :user_work_tags, :work_id
    add_index :user_work_tags, :work_tag_id
    add_index :user_work_tags, %i(user_id work_id work_tag_id), unique: true
    add_foreign_key :user_work_tags, :users
    add_foreign_key :user_work_tags, :works
    add_foreign_key :user_work_tags, :work_tags
  end
end
