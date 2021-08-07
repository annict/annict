# frozen_string_literal: true

class CreateFavoriteCharacters < ActiveRecord::Migration[5.0]
  def change
    create_table :favorite_characters do |t|
      t.integer :user_id, null: false
      t.integer :character_id, null: false
      t.timestamps null: false
    end

    add_index :favorite_characters, :user_id
    add_index :favorite_characters, :character_id
    add_index :favorite_characters, %i[user_id character_id], unique: true

    add_foreign_key :favorite_characters, :users
    add_foreign_key :favorite_characters, :characters

    add_column :characters, :favorite_characters_count, :integer, null: false, default: 0
    add_index :characters, :favorite_characters_count
  end
end
