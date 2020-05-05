# frozen_string_literal: true

class ChangeSchema202005 < ActiveRecord::Migration[6.0]
  def change
    add_column :users, :character_favorites_count, :integer, null: false, default: 0
    add_column :users, :person_favorites_count, :integer, null: false, default: 0
    add_column :users, :organization_favorites_count, :integer, null: false, default: 0
  end
end
