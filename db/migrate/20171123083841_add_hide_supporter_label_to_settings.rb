# frozen_string_literal: true

class AddHideSupporterLabelToSettings < ActiveRecord::Migration[5.1]
  def change
    add_column :settings, :hide_supporter_label, :boolean, default: false, null: false
  end
end
