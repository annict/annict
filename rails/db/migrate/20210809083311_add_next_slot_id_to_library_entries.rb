# frozen_string_literal: true

class AddNextSlotIdToLibraryEntries < ActiveRecord::Migration[6.1]
  def change
    add_column :library_entries, :next_slot_id, :bigint
    add_index :library_entries, :next_slot_id
    add_foreign_key :library_entries, :slots, column: :next_slot_id
  end
end
