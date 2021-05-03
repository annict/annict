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

    add_column :users, :following_count, :integer, null: false, default: 0
    add_column :users, :followers_count, :integer, null: false, default: 0

    rename_column :works, :auto_episodes_count, :episodes_count

    add_column :activities, :activity_group_id, :bigint
    add_column :activities, :migrated_at, :datetime
    add_column :activities, :mer_processed_at, :datetime

    add_index :activities, :activity_group_id
    add_index :activities, :created_at
    add_index :activities, %i[activity_group_id created_at]
    add_index :activities, %i[trackable_id trackable_type]

    add_foreign_key :activities, :activity_groups
  end
end
