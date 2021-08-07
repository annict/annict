# frozen_string_literal: true

class UpdateTips < ActiveRecord::Migration[5.1]
  def change
    remove_index :tips, name: :index_tips_on_slug
    add_index :tips, %i[slug locale], unique: true

    remove_column :tips, :title_en
  end
end
