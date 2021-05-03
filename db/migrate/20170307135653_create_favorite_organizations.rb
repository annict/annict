# frozen_string_literal: true

class CreateFavoriteOrganizations < ActiveRecord::Migration[5.0]
  def change
    create_table :favorite_organizations do |t|
      t.integer :user_id, null: false
      t.integer :organization_id, null: false
      t.timestamps null: false
    end

    add_index :favorite_organizations, :user_id
    add_index :favorite_organizations, :organization_id
    add_index :favorite_organizations, %i[user_id organization_id], unique: true

    add_foreign_key :favorite_organizations, :users
    add_foreign_key :favorite_organizations, :organizations

    add_column :organizations, :favorite_organizations_count, :integer, null: false, default: 0
    add_index :organizations, :favorite_organizations_count
  end
end
