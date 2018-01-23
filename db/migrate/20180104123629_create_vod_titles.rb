# frozen_string_literal: true

class CreateVodTitles < ActiveRecord::Migration[5.1]
  def change
    create_table :vod_titles do |t|
      t.references :channel, null: false, foreign_key: true
      t.references :work, foreign_key: true
      t.string :code, null: false
      t.string :name, null: false
      t.string :aasm_state, null: false, default: "published"
      t.datetime :mail_sent_at
      t.timestamps null: false
    end

    add_index :vod_titles, :mail_sent_at
  end
end
