# frozen_string_literal: true

class AddFormatToNumberFormats < ActiveRecord::Migration
  def change
    add_column :number_formats, :format, :string, null: false, default: ""
  end
end
