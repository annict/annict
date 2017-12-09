# frozen_string_literal: true

class RenameHideSupporterLabelToHideSupporterBadge < ActiveRecord::Migration[5.1]
  def change
    rename_column :settings, :hide_supporter_label, :hide_supporter_badge
  end
end
