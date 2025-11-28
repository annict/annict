# frozen_string_literal: true

class RenameCheckinsToRecords < ActiveRecord::Migration[5.1]
  def change
    rename_table :checkins, :records
    rename_column :comments, :checkin_id, :record_id
    rename_column :settings, :hide_checkin_comment, :hide_record_comment
    rename_column :episodes, :checkins_count, :records_count
    rename_column :users, :checkins_count, :records_count
  end
end
