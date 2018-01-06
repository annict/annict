# frozen_string_literal: true

class AddVodTitleCodeToProgramDetails < ActiveRecord::Migration[5.1]
  def change
    add_column :program_details, :vod_title_code, :string, null: false, default: ""
    add_index :program_details, :vod_title_code
  end
end
