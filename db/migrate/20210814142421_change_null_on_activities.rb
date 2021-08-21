# frozen_string_literal: true

class ChangeNullOnActivities < ActiveRecord::Migration[6.1]
  def change
    change_column_null :activities, :trackable_id, true
    change_column_null :activities, :trackable_type, true
  end
end
