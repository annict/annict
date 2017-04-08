# frozen_string_literal: true

class CreateProjectMembers < ActiveRecord::Migration[5.0]
  def change
    create_table :project_members do |t|
      t.integer :user_id, null: false
      t.integer :project_id, null: false
      t.timestamps null: false
    end

    add_index :project_members, :user_id
    add_index :project_members, :project_id
    add_index :project_members, %i(user_id project_id), unique: true

    add_foreign_key :project_members, :users
    add_foreign_key :project_members, :projects
  end
end
