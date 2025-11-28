# frozen_string_literal: true

class ChangeDefaultValueOnDisplayOptionUserWorkList < ActiveRecord::Migration[5.0]
  def change
    change_column_default :settings, :display_option_user_work_list, :grid_detailed
    change_column_default :settings, :display_option_work_list, :list_detailed
  end
end
