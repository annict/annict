# frozen_string_literal: true

class AddKeyPvToWorks < ActiveRecord::Migration[5.1]
  def change
    add_column :works, :key_pv_id, :integer
    add_index :works, :key_pv_id
    add_foreign_key :works, :pvs, column: :key_pv_id
  end
end
