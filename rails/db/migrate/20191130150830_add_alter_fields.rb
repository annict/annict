# frozen_string_literal: true

class AddAlterFields < ActiveRecord::Migration[6.0]
  def change
    add_column :channels, :name_alter, :string, null: false, default: ""
    add_column :works, :title_alter, :string, null: false, default: ""
    add_column :works, :title_alter_en, :string, null: false, default: ""
  end
end
