# frozen_string_literal: true

class AddNameEnCasts < ActiveRecord::Migration[5.0]
  def change
    add_column :casts, :name_en, :string, null: false, default: ""
  end
end
