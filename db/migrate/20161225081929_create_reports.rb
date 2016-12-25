# frozen_string_literal: true

class CreateReports < ActiveRecord::Migration[5.0]
  def change
    create_table :reports do |t|
      t.integer :user_id, null: false
      t.string :root_resource_type, null: false
      t.integer :root_resource_id, null: false
      t.string :resource_type
      t.integer :resource_id
      t.timestamps null: false
    end

    add_index :reports, :user_id
    add_index :reports, %i(root_resource_id root_resource_type)
    add_index :reports, %i(resource_id resource_type)

    add_foreign_key :reports, :users
  end
end
