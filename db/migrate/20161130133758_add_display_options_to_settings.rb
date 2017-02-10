# frozen_string_literal: true

class AddDisplayOptionsToSettings < ActiveRecord::Migration[5.0]
  def change
    add_column :settings, :display_option_work_list, :string,
      null: false,
      default: "list"
    add_column :settings, :display_option_user_work_list, :string,
      null: false,
      default: "list"
  end
end
