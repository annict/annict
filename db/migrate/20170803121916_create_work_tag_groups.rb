# frozen_string_literal: true

class CreateWorkTagGroups < ActiveRecord::Migration[5.1]
  def change
    create_table :work_tag_groups do |t|
      t.string :name, null: false
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end
  end
end
