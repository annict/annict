# frozen_string_literal: true

class AddProgramIdToLibraryEntries < ActiveRecord::Migration[6.0]
  def change
    add_column :library_entries, :program_id, :bigint

    add_index :library_entries, :program_id
    add_index :library_entries, %i[user_id program_id], unique: true

    add_foreign_key :library_entries, :programs
  end
end
