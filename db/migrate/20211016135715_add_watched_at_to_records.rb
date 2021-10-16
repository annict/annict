# frozen_string_literal: true

class AddWatchedAtToRecords < ActiveRecord::Migration[6.1]
  def change
    add_column :records, :watched_at, :datetime
    add_index :records, :watched_at
  end
end
