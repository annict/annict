# frozen_string_literal: true

class AddProgramsSortTypeToSettings < ActiveRecord::Migration[5.0]
  def change
    add_column :settings, :programs_sort_type, :string, null: false, default: ""
  end
end
