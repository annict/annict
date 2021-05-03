# frozen_string_literal: true

class CreateUserlandProjectMembers < ActiveRecord::Migration[5.0]
  def change
    create_table :userland_project_members do |t|
      t.integer :user_id, null: false
      t.integer :userland_project_id, null: false
      t.timestamps null: false
    end

    add_index :userland_project_members, :user_id
    add_index :userland_project_members, :userland_project_id
    add_index :userland_project_members, %i[user_id userland_project_id],
      unique: true,
      name: :index_userland_pm_on_uid_and_userland_pid

    add_foreign_key :userland_project_members, :users
    add_foreign_key :userland_project_members, :userland_projects
  end
end
