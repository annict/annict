# frozen_string_literal: true

class RemoveNullConstInPersonIdOnStaffs < ActiveRecord::Migration[4.2]
  def change
    change_column_null :staffs, :person_id, true
    change_column_null :draft_staffs, :person_id, true
  end
end
