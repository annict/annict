# frozen_string_literal: true

class AddCharacterIdToCasts < ActiveRecord::Migration[5.0]
  def change
    add_column :casts, :character_id, :integer
    add_index :casts, :character_id
    add_index :casts, [:work_id, :character_id], unique: true
    add_foreign_key :casts, :characters
  end
end
