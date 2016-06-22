# frozen_string_literal: true

class AddImageColorToItems < ActiveRecord::Migration[5.0]
  def change
    add_column :items, :image_color_light, :string, null: false, default: ""
    add_column :items, :image_color_dark, :string, null: false, default: ""
    add_column :draft_items, :image_color_light, :string, null: false, default: ""
    add_column :draft_items, :image_color_dark, :string, null: false, default: ""
  end
end
