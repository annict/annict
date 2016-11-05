# frozen_string_literal: true

class AddTitleEnEtcToTips < ActiveRecord::Migration[5.0]
  def change
    add_column :tips, :title_en, :string, null: false, default: ""
    add_column :tips, :body, :text, null: false, default: ""
    add_column :tips, :body_en, :text, null: false, default: ""
    rename_column :tips, :partial_name, :slug
  end
end
