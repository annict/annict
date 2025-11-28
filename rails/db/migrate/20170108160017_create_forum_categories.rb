# frozen_string_literal: true

class CreateForumCategories < ActiveRecord::Migration[5.0]
  def change
    create_table :forum_categories do |t|
      t.string :slug, null: false
      t.string :name, null: false
      t.string :name_en, null: false
      t.string :description, null: false
      t.string :description_en, null: false
      t.string :postable_role, null: false
      t.integer :sort_number, null: false
      t.integer :forum_posts_count, null: false, default: 0
      t.timestamps null: false
    end

    add_index :forum_categories, :slug, unique: true
  end
end
