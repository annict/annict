# frozen_string_literal: true

class AddSessionsTable < ActiveRecord::Migration[5.2]
  def change
    drop_table :sessions
    create_table :sessions do |t|
      t.string :session_id, null: false
      t.jsonb :data, null: false, default: "{}"
      t.timestamps null: false
    end

    add_index :sessions, :session_id, unique: true
    add_index :sessions, :updated_at
  end
end
