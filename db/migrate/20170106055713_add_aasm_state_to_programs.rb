# frozen_string_literal: true

class AddAasmStateToPrograms < ActiveRecord::Migration[5.0]
  def change
    add_column :programs, :aasm_state, :string, null: false, default: "published"
    add_index :programs, :aasm_state
  end
end
