# frozen_string_literal: true

class CreateMultipleRecords < ActiveRecord::Migration[4.2]
  def change
    create_table :multiple_records do |t|
      t.integer :user_id, null: false
      t.integer :work_id, null: false
      t.integer :likes_count, null: false, default: 0
      t.timestamps null: false
    end

    add_index :multiple_records, :user_id
    add_foreign_key :multiple_records, :users
    add_index :multiple_records, :work_id
    add_foreign_key :multiple_records, :works

    add_column :checkins, :multiple_record_id, :integer
    add_index :checkins, :multiple_record_id
    add_foreign_key :checkins, :multiple_records
  end
end
