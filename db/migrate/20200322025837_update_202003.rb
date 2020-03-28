# frozen_string_literal: true

class Update202003 < ActiveRecord::Migration[6.0]
  def change
    add_column :series, :name_alter, :string, null: false, default: ""
    add_column :series, :name_alter_en, :string, null: false, default: ""
  end
end
