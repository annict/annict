# frozen_string_literal: true

class CreateWorkTags < ActiveRecord::Migration[5.1]
  def change
    create_table :work_tags do |t|
      t.integer :work_tag_group_id, null: false
      t.string :name, null: false
      t.string :aasm_state, null: false, default: "published"
      t.integer :user_work_tags_count, null: false, default: 0
      t.timestamps null: false
    end

    add_index :work_tags, :work_tag_group_id
    add_foreign_key :work_tags, :work_tag_groups
  end
end
