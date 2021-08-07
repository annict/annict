# frozen_string_literal: true

class CreateFavoritePeople < ActiveRecord::Migration[5.0]
  def change
    create_table :favorite_people do |t|
      t.integer :user_id, null: false
      t.integer :person_id, null: false
      t.timestamps null: false
    end

    add_index :favorite_people, :user_id
    add_index :favorite_people, :person_id
    add_index :favorite_people, %i[user_id person_id], unique: true

    add_foreign_key :favorite_people, :users
    add_foreign_key :favorite_people, :people

    add_column :people, :favorite_people_count, :integer, null: false, default: 0
    add_index :people, :favorite_people_count
  end
end
