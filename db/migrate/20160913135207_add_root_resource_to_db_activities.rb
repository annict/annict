# frozen_string_literal: true

class AddRootResourceToDbActivities < ActiveRecord::Migration[5.0]
  def change
    add_column :db_activities, :root_resource_id, :integer
    add_column :db_activities, :root_resource_type, :string
    add_index :db_activities, [:root_resource_id, :root_resource_type]
  end
end
