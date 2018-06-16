# frozen_string_literal: true

class RebuildRecords < ActiveRecord::Migration[5.1]
  def change
    rename_table :multiple_records, :multiple_episode_records
    rename_table :records, :episode_records
    rename_table :reviews, :work_records
    rename_table :new_records, :records

    rename_column :activities, :multiple_record_id, :multiple_episode_record_id
    rename_column :activities, :record_id, :episode_record_id
    rename_column :activities, :review_id, :work_record_id
    rename_column :comments, :record_id, :episode_record_id
    rename_column :email_notifications, :event_liked_record, :event_liked_episode_record
    rename_column :episode_records, :multiple_record_id, :multiple_episode_record_id
    rename_column :episode_records, :new_record_id, :record_id
    rename_column :episodes, :record_comments_count, :episode_records_with_body_count
    rename_column :episodes, :records_count, :episode_records_count
    rename_column :users, :records_count, :episode_records_count
    rename_column :works, :reviews_count, :work_records_count
    rename_column :work_records, :new_record_id, :record_id

    add_column :works, :records_count, :integer, null: false, default: 0
    add_column :works, :work_records_with_body_count, :integer, null: false, default: 0
    add_column :users, :records_count, :integer, null: false, default: 0
  end
end
