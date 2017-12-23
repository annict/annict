class CreateLatestStatuses < ActiveRecord::Migration[4.2]
  def change
    create_table :latest_statuses do |t|
      t.integer :user_id, null: false
      t.integer :work_id, null: false
      t.integer :next_episode_id
      t.integer :kind, null: false
      t.integer :watched_episode_ids, default: [], null: false, array: true
      t.integer :position, default: 0, null: false
      t.timestamps null: false
    end

    add_index :latest_statuses, :user_id
    add_index :latest_statuses, :work_id
    add_index :latest_statuses, :next_episode_id
    add_index :latest_statuses, [:user_id, :work_id], unique: true
    add_index :latest_statuses, [:user_id, :position]

    add_foreign_key :latest_statuses, :users
    add_foreign_key :latest_statuses, :works
    add_foreign_key :latest_statuses, :episodes, column: :next_episode_id
  end
end
