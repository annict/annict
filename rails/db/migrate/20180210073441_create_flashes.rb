# frozen_string_literal: true

class CreateFlashes < ActiveRecord::Migration[5.1]
  def change
    create_table :flashes do |t|
      t.string :client_uuid, null: false
      t.json :data
      t.timestamps null: false
    end

    add_index :flashes, :client_uuid, unique: true
  end
end
