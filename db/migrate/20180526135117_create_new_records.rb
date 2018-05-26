# frozen_string_literal: true

class CreateNewRecords < ActiveRecord::Migration[5.1]
  def change
    create_table :new_records do |t|
      t.references :user, null: false, foreign_key: true
      t.references :work, null: false, foreign_key: true
      t.string :aasm_state, null: false, default: "published"
      t.integer :impressions_count, null: false, default: 0
      t.timestamps null: false
    end

    add_column :records, :new_record_id, :integer
    add_column :reviews, :new_record_id, :integer

    add_index :records, :new_record_id, unique: true
    add_index :reviews, :new_record_id, unique: true

    add_foreign_key :records, :new_records
    add_foreign_key :reviews, :new_records
  end
end
