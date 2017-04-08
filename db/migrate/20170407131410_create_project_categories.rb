# frozen_string_literal: true

class CreateProjectCategories < ActiveRecord::Migration[5.0]
  def change
    create_table :project_categories do |t|
      t.string :name, null: false
      t.string :name_en, null: false
      t.integer :projects_count, null: false, default: 0
      t.timestamps null: false
    end
  end
end
