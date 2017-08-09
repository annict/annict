# frozen_string_literal: true

class CreateWorkTags < ActiveRecord::Migration[5.1]
  def change
    create_table :work_tags do |t|
      t.integer :work_tag_group_id, null: false
      t.integer :user_id, null: false
      t.string :name, null: false
      t.string :description
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :work_tags, :work_tag_group_id
    add_foreign_key :work_tags, :work_tag_groups

    add_index :work_tags, :user_id
    add_foreign_key :work_tags, :users

    add_index :work_tags, :name

    add_index :work_tags, %i(user_id name), unique: true
  end
end
