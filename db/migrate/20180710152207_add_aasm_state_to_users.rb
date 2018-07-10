# frozen_string_literal: true

class AddAasmStateToUsers < ActiveRecord::Migration[5.2]
  def change
    add_column :users, :aasm_state, :string, null: false, default: "published"
    add_index :users, :aasm_state
  end
end
