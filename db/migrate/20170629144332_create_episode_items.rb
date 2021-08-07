# frozen_string_literal: true

class CreateEpisodeItems < ActiveRecord::Migration[5.1]
  def change
    create_table :episode_items do |t|
      t.integer :episode_id, null: false
      t.integer :item_id, null: false
      t.integer :user_id, null: false
      t.integer :work_id, null: false
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :episode_items, :episode_id
    add_index :episode_items, :item_id
    add_index :episode_items, :user_id
    add_index :episode_items, :work_id
    add_index :episode_items, %i[episode_id item_id], unique: true

    add_foreign_key :episode_items, :episodes
    add_foreign_key :episode_items, :items
    add_foreign_key :episode_items, :users
    add_foreign_key :episode_items, :works
  end
end
