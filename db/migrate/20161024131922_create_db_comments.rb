# frozen_string_literal: true

class CreateDbComments < ActiveRecord::Migration[5.0]
  def change
    create_table :db_comments do |t|
      t.integer :user_id, null: false
      t.integer :resource_id, null: false
      t.string :resource_type, null: false
      t.text :body, null: false
      t.timestamps
    end

    add_index :db_comments, :user_id
    add_index :db_comments, %i(resource_id resource_type)
  end
end
