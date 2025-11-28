# frozen_string_literal: true

class AddVodTitleCodeToProgramDetails < ActiveRecord::Migration[5.1]
  def change
    add_column :program_details, :vod_title_code, :string, null: false, default: ""
    add_index :program_details, :vod_title_code

    add_column :program_details, :vod_title_name, :string, null: false, default: ""

    remove_index :program_details, name: :index_program_details_on_channel_id_and_work_id
  end
end
