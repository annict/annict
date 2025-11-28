# frozen_string_literal: true

class AddCastsCountAndStaffsCountToPeopleAndOrgs < ActiveRecord::Migration[5.0]
  def change
    add_column :people, :casts_count, :integer, null: false, default: 0
    add_column :people, :staffs_count, :integer, null: false, default: 0
    add_column :organizations, :staffs_count, :integer, null: false, default: 0

    add_index :people, :casts_count
    add_index :people, :staffs_count
    add_index :organizations, :staffs_count
  end
end
