# frozen_string_literal: true

class AddSeriesIdToCharacters < ActiveRecord::Migration[5.0]
  def change
    add_column :characters, :series_id, :integer

    add_index :characters, :series_id
    add_index :characters, %i(name series_id), unique: true

    add_foreign_key_constraint :characters, :series, on_delete: :set_null

    remove_column :characters, :kind_en

    remove_index :characters, %i(name kind)
  end
end
