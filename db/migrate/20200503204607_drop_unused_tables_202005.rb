# frozen_string_literal: true

class DropUnusedTables202005 < ActiveRecord::Migration[6.0]
  def change
    drop_table :impressions
    drop_table :episode_items
    drop_table :work_items
    drop_table :items
  end
end
