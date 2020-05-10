# frozen_string_literal: true

class UpdateProfileV4Pre2 < ActiveRecord::Migration[6.0]
  def change
    change_column_null :activities, :recipient_id, true
    change_column_null :activities, :recipient_type, true
    change_column_null :activities, :trackable_id, true
  end
end
