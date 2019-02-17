# frozen_string_literal: true

class UpdateForManualPrograms < ActiveRecord::Migration[5.2]
  def change
    add_column :program_details, :rebroadcast, :boolean, default: false, null: false
    change_column_null :programs, :episode_id, true
  end
end
