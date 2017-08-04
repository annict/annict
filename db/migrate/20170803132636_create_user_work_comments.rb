# frozen_string_literal: true

class CreateUserWorkComments < ActiveRecord::Migration[5.1]
  def change
    create_table :user_work_comments do |t|
      t.integer :user_id, null: false
      t.integer :work_id, null: false
      t.string :body, null: false
      t.timestamps null: false
    end

    add_index :user_work_comments, :user_id
    add_index :user_work_comments, :work_id
    add_index :user_work_comments, %i(user_id work_id), unique: true
    add_foreign_key :user_work_comments, :users
    add_foreign_key :user_work_comments, :works
  end
end
