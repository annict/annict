# frozen_string_literal: true

class ChangeSchema202005 < ActiveRecord::Migration[6.0]
  def change
    add_column :users, :favorite_characters_count, :integer, null: false, default: 0
    add_column :users, :favorite_people_count, :integer, null: false, default: 0
    add_column :users, :favorite_organizations_count, :integer, null: false, default: 0
  end
end
