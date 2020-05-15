# frozen_string_literal: true

class UpdateActivity202005 < ActiveRecord::Migration[6.0]
  def change
    change_column_null :activities, :recipient_id, true
    change_column_null :activities, :recipient_type, true
    change_column_null :activities, :action, true
  end
end
