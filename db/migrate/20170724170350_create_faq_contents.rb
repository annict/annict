# frozen_string_literal: true

class CreateFaqContents < ActiveRecord::Migration[5.1]
  def change
    create_table :faq_contents do |t|
      t.integer :faq_category_id, null: false
      t.string :question, null: false
      t.text :answer, null: false
      t.string :locale, null: false
      t.integer :sort_number, null: false, default: 0
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :faq_contents, :faq_category_id
    add_index :faq_contents, :locale

    add_foreign_key :faq_contents, :faq_categories
  end
end
