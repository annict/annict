# frozen_string_literal: true

class CreateGuests < ActiveRecord::Migration[5.2]
  def change
    create_table :guests do |t|
      t.string :uuid, null: false
      t.string :user_agent, null: false, default: ""
      t.string :remote_ip, null: false, default: ""
      t.string :time_zone, null: false
      t.string :locale, null: false
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :guests, :uuid, unique: true
  end
end
