# frozen_string_literal: true

class RenameLatestStatusesToLibraryEntries < ActiveRecord::Migration[6.0]
  def change
    rename_table :latest_statuses, :library_entries

    change_column_null :library_entries, :kind, true
  end
end
