# frozen_string_literal: true

class AddItemableToActivities < ActiveRecord::Migration[6.1]
  def change
    add_column :activities, :itemable_id, :bigint
    add_column :activities, :itemable_type, :string
    add_index :activities, %i[itemable_id itemable_type]
  end
end
