# frozen_string_literal: true

class ChangeSchemaForUserDetail202005 < ActiveRecord::Migration[6.0]
  def change
    add_column :users, :character_favorites_count, :integer, null: false, default: 0
    add_column :users, :person_favorites_count, :integer, null: false, default: 0
    add_column :users, :organization_favorites_count, :integer, null: false, default: 0

    add_column :users, :plan_to_watch_works_count, :integer, null: false, default: 0
    add_column :users, :watching_works_count, :integer, null: false, default: 0
    add_column :users, :completed_works_count, :integer, null: false, default: 0
    add_column :users, :on_hold_works_count, :integer, null: false, default: 0
    add_column :users, :dropped_works_count, :integer, null: false, default: 0
  end
end
