# frozen_string_literal: true

class AddTitleEnToPvs < ActiveRecord::Migration[5.1]
  def change
    add_column :pvs, :title_en, :string, null: false, default: ""
  end
end
