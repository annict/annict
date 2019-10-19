# frozen_string_literal: true

class RemoveIndexSlotsOnProgramIdAndStartedAt < ActiveRecord::Migration[6.0]
  def change
    remove_index :slots, name: :index_slots_on_program_id_and_started_at
  end
end
