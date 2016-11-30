# frozen_string_literal: true

class RenameRecipientToObjectOnDbActivities < ActiveRecord::Migration[5.0]
  def change
    remove_column :db_activities, :recipient_id
    remove_column :db_activities, :recipient_type

    add_column :db_activities, :object_id, :integer
    add_column :db_activities, :object_type, :string

    add_index :db_activities, %i(object_id object_type)
  end
end
