# frozen_string_literal: true

class AddResourceIdToDraftStaffs < ActiveRecord::Migration[4.2]
  def change
    add_column :draft_staffs, :resource_id, :integer
    add_column :draft_staffs, :resource_type, :string

    add_index :draft_staffs, [:resource_id, :resource_type]
  end
end
