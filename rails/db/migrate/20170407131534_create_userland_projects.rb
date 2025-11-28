# frozen_string_literal: true

class CreateUserlandProjects < ActiveRecord::Migration[5.0]
  def change
    create_table :userland_projects do |t|
      t.integer :userland_category_id, null: false
      t.string :name, null: false
      t.string :summary, null: false
      t.text :description, null: false
      t.string :url, null: false
      t.attachment :icon
      t.boolean :available, null: false, default: false
      t.timestamps null: false
    end

    add_index :userland_projects, :userland_category_id

    add_foreign_key :userland_projects, :userland_categories
  end
end
