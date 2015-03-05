class CreateChecks < ActiveRecord::Migration
  def change
    create_table :checks do |t|
      t.integer :user_id,    null: false
      t.integer :work_id,    null: false
      t.integer :episode_id
      t.integer :skipped_episode_ids, array: true, default: [], null: false
      t.integer :position, default: 0, null: false
      t.timestamps null: false
    end

    add_index :checks, :user_id
    add_index :checks, [:user_id, :work_id], unique: true
    add_index :checks, [:user_id, :position]

    add_foreign_key :checks, :users
    add_foreign_key :checks, :works
    add_foreign_key :checks, :episodes
  end
end
