# frozen_string_literal: true

class CreateFavorites < ActiveRecord::Migration[5.0]
  def change
    create_table :favorites do |t|
      t.integer :user_id, null: false
      t.string :resource_type, null: false
      t.integer :resource_id, null: false
      t.timestamps null: false
    end

    add_index :favorites, :user_id
    add_index :favorites, %i(user_id resource_type resource_id), unique: true

    add_foreign_key :favorites, :users

    add_column :people, :favorites_count, :integer, null: false, default: 0
    add_column :organizations, :favorites_count, :integer, null: false, default: 0
    add_column :characters, :favorites_count, :integer, null: false, default: 0
  end
end
