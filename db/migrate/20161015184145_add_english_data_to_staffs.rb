# frozen_string_literal: true

class AddEnglishDataToStaffs < ActiveRecord::Migration[5.0]
  def change
    add_column :staffs, :name_en, :string, null: false, default: ""
    add_column :staffs, :role_other_en, :string, null: false, default: ""
  end
end
