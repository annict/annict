# frozen_string_literal: true

class CreateEmailConfirmations < ActiveRecord::Migration[6.0]
  def change
    enable_extension "citext"

    create_table :email_confirmations do |t|
      t.citext :email, null: false
      t.string :event, null: false
      t.string :token, null: false
      t.string :back
      t.datetime :expires_at, null: false
      t.timestamps
    end

    add_index :email_confirmations, :token, unique: true
  end
end
