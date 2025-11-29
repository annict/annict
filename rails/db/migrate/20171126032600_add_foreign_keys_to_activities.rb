# frozen_string_literal: true

class AddForeignKeysToActivities < ActiveRecord::Migration[5.1]
  def change
    add_column :activities, :work_id, :integer
    add_index :activities, :work_id
    add_foreign_key :activities, :works

    add_column :activities, :episode_id, :integer
    add_index :activities, :episode_id
    add_foreign_key :activities, :episodes

    add_column :activities, :status_id, :integer
    add_index :activities, :status_id
    add_foreign_key :activities, :statuses

    add_column :activities, :record_id, :integer
    add_index :activities, :record_id
    add_foreign_key :activities, :checkins, column: :record_id

    add_column :activities, :multiple_record_id, :integer
    add_index :activities, :multiple_record_id
    add_foreign_key :activities, :multiple_records

    add_column :activities, :review_id, :integer
    add_index :activities, :review_id
    add_foreign_key :activities, :reviews
  end
end
