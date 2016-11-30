# frozen_string_literal: true

class CreateDbComments < ActiveRecord::Migration[5.0]
  def change
    create_table :db_comments do |t|
      t.text :body, null: false
      t.timestamps
    end
  end
end
