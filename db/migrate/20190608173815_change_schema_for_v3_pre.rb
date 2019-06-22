# frozen_string_literal: true
# typed: false

class ChangeSchemaForV3Pre < ActiveRecord::Migration[5.2]
  def change
    add_column :users, :deleted_at, :datetime

    add_index :users, :deleted_at
  end
end
