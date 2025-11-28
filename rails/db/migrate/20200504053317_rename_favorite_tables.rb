# frozen_string_literal: true

class RenameFavoriteTables < ActiveRecord::Migration[6.0]
  def change
    rename_table :favorite_characters, :character_favorites
    rename_table :favorite_organizations, :organization_favorites
    rename_table :favorite_people, :person_favorites
  end
end
