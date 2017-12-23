# frozen_string_literal: true

class AddFormatToNumberFormats < ActiveRecord::Migration[4.2]
  def change
    add_column :number_formats, :format, :string, null: false, default: ""
  end
end
