# frozen_string_literal: true

class CreateStreamingLinks < ActiveRecord::Migration[5.1]
  def change
    create_table :streaming_links do |t|
      t.integer :channel_id, null: false
      t.integer :work_id, null: false
      t.string :locale, null: false
      t.string :unique_id, null: false
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :streaming_links, :channel_id
    add_index :streaming_links, :work_id
    add_index :streaming_links, %i(channel_id work_id locale), unique: true
    add_index :streaming_links, %i(channel_id locale unique_id), unique: true

    add_foreign_key :streaming_links, :channels
    add_foreign_key :streaming_links, :works
  end
end
