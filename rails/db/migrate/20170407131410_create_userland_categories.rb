# frozen_string_literal: true

class CreateUserlandCategories < ActiveRecord::Migration[5.0]
  def change
    create_table :userland_categories do |t|
      t.string :name, null: false
      t.string :name_en, null: false
      t.integer :sort_number, null: false, default: 0
      t.integer :userland_projects_count, null: false, default: 0
      t.timestamps null: false
    end
  end
end
