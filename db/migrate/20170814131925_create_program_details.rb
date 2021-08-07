# frozen_string_literal: true

class CreateProgramDetails < ActiveRecord::Migration[5.1]
  def change
    create_table :program_details do |t|
      t.integer :channel_id, null: false
      t.integer :work_id, null: false
      t.string :url
      t.datetime :started_at
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :program_details, :channel_id
    add_index :program_details, :work_id
    add_index :program_details, %i[channel_id work_id], unique: true

    add_foreign_key :program_details, :channels
    add_foreign_key :program_details, :works

    add_column :channels, :video_service, :boolean, default: false
    add_column :channels, :aasm_state, :string, null: false, default: "published"

    add_index :channels, :video_service

    change_column_null :channels, :sc_chid, true

    remove_column :channels, :published
  end
end
