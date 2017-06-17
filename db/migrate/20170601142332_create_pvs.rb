# frozen_string_literal: true

class CreatePvs < ActiveRecord::Migration[5.1]
  def change
    create_table :pvs do |t|
      t.integer :work_id, null: false
      t.string :url, null: false
      t.string :title, null: false
      t.attachment :thumbnail
      t.integer :sort_number, null: false, default: 0
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :pvs, :work_id
    add_foreign_key :pvs, :works
  end
end
