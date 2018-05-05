# frozen_string_literal: true

class CreateRecords < ActiveRecord::Migration[5.1]
  def change
    rename_table :multiple_records, :multiple_episode_records
    rename_table :records, :episode_records
    rename_table :reviews, :work_records

    rename_column :activities, :multiple_record_id, :multiple_episode_record_id
    rename_column :activities, :record_id, :episode_record_id
    rename_column :activities, :review_id, :work_record_id
    rename_column :episode_records, :multiple_record_id, :multiple_episode_record_id
    rename_column :episodes, :record_comments_count, :episode_record_comments_count
    rename_column :episodes, :records_count, :episode_records_count

    create_table :records do |t|
      t.references :user, null: false, foreign_key: true
      t.timestamps null: false
    end

    add_column :episode_records, :record_id, :integer
    add_column :work_records, :record_id, :integer

    add_index :episode_records, :record_id, unique: true
    add_index :work_records, :record_id, unique: true

    add_foreign_key :episode_records, :records
    add_foreign_key :work_records, :records
  end
end
