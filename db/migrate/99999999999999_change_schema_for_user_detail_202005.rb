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

    add_column :statuses, :activity_id, :bigint
    add_column :episode_records, :activity_id, :bigint
    add_column :work_records, :activity_id, :bigint

    add_index :statuses, :activity_id
    add_index :episode_records, :activity_id
    add_index :work_records, :activity_id

    add_foreign_key :statuses, :activities
    add_foreign_key :episode_records, :activities
    add_foreign_key :work_records, :activities

    add_column :activities, :resources_count, :integer, null: false, default: 0
    add_column :activities, :single, :boolean, null: false, default: false
    add_column :activities, :repetitiveness, :boolean, null: false, default: false

    add_index :activities, :repetitiveness
  end
end
