# frozen_string_literal: true

class Update202008Pre < ActiveRecord::Migration[6.0]
  def change
    add_column :statuses, :new_kind, :string
  end
end
