# frozen_string_literal: true

class UpdateColumns20160305 < ActiveRecord::Migration[4.2]
  def change
    remove_column :staffs, :person_id
    remove_column :draft_staffs, :person_id
    change_column_null :staffs, :resource_id, false
    change_column_null :staffs, :resource_type, false
  end
end
