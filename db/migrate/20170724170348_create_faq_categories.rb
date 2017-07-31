# frozen_string_literal: true

class CreateFaqCategories < ActiveRecord::Migration[5.1]
  def change
    create_table :faq_categories do |t|
      t.string :name, null: false
      t.string :locale, null: false
      t.integer :sort_number, null: false, default: 0
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :faq_categories, :locale
  end
end
