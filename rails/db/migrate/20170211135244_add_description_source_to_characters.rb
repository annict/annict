# frozen_string_literal: true

class AddDescriptionSourceToCharacters < ActiveRecord::Migration[5.0]
  def change
    add_column :characters, :description_source, :string, null: false, default: ""
    add_column :characters, :description_source_en, :string, null: false, default: ""
  end
end
