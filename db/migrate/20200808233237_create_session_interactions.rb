# frozen_string_literal: true

class CreateSessionInteractions < ActiveRecord::Migration[6.0]
  def change
    enable_extension "citext"

    create_table :session_interactions do |t|
      t.citext :email, null: false
      t.string :kind, null: false
      t.string :token, null: false
      t.datetime :expires_at, null: false
      t.timestamps
    end

    add_index :session_interactions, :email, unique: true
    add_index :session_interactions, :token, unique: true
  end
end
