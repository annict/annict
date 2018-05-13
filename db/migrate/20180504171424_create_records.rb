# frozen_string_literal: true

class CreateRecords < ActiveRecord::Migration[5.1]
  def change
    rename_table :multiple_records, :multiple_episode_records
    rename_table :records, :episode_records
    rename_table :reviews, :work_records

    rename_column :activities, :multiple_record_id, :multiple_episode_record_id
    rename_column :activities, :record_id, :episode_record_id
    rename_column :activities, :review_id, :work_record_id
    rename_column :comments, :record_id, :episode_record_id
    rename_column :episode_records, :multiple_record_id, :multiple_episode_record_id
    rename_column :episodes, :record_comments_count, :episode_record_comments_count
    rename_column :episodes, :records_count, :episode_records_count
    rename_column :users, :records_count, :episode_records_count
    rename_column :works, :reviews_count, :work_records_count

    create_table :records do |t|
      t.references :user, null: false, foreign_key: true
      t.references :work, null: false, foreign_key: true
      t.string :aasm_state, null: false, default: "published"
      t.integer :impressions_count, null: false, default: 0
      t.timestamps null: false
    end

    add_column :episode_records, :record_id, :integer
    add_column :work_records, :record_id, :integer
    add_column :works, :records_count, :integer, null: false, default: 0
    add_column :users, :records_count, :integer, null: false, default: 0

    add_index :episode_records, :record_id, unique: true
    add_index :work_records, :record_id, unique: true

    add_foreign_key :episode_records, :records
    add_foreign_key :work_records, :records
  end
end
