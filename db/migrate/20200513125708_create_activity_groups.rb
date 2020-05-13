# frozen_string_literal: true

class CreateActivityGroups < ActiveRecord::Migration[6.0]
  def change
    create_table :activity_groups do |t|
      t.string :action, null: false
      t.boolean :single, null: false, default: false
      t.integer :activities_count, null: false, default: 0
      t.timestamps null: false
    end

    add_index :activity_groups, :created_at
  end
end
