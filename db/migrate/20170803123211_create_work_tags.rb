# frozen_string_literal: true

class CreateWorkTags < ActiveRecord::Migration[5.1]
  def change
    create_table :work_tags do |t|
      t.string :name, null: false
      t.string :aasm_state, null: false, default: "published"
      t.integer :work_taggings_count, null: false, default: 0
      t.timestamps null: false
    end

    add_index :work_tags, :name, unique: true
    add_index :work_tags, :work_taggings_count
  end
end
