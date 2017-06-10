# frozen_string_literal: true

class AddRecordsSortTypeToSettings < ActiveRecord::Migration[5.1]
  def change
    add_column :settings, :records_sort_type, :string, null: false, default: "created_at_desc"
    add_column :settings, :display_option_record_list, :string, null: false, default: "all_comments"
  end
end
