# frozen_string_literal: true

class CreateActivityGroups < ActiveRecord::Migration[6.0]
  def change
    create_table :activity_groups do |t|
      t.bigint :user_id, null: false
      t.string :itemable_type, null: false
      t.boolean :single, null: false, default: false
      t.integer :activities_count, null: false, default: 0
      t.timestamps null: false
    end

    add_index :activity_groups, :user_id
    add_index :activity_groups, :created_at

    add_foreign_key :activity_groups, :users
  end
end
