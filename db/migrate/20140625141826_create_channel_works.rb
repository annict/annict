class CreateChannelWorks < ActiveRecord::Migration
  def change
    create_table :channel_works do |t|
      t.integer :user_id,    null: false
      t.integer :work_id,    null: false
      t.integer :channel_id, null: false
      t.timestamps
    end

    add_index :channel_works, [:user_id, :work_id]
    add_index :channel_works, [:user_id, :work_id, :channel_id], unique: true

    add_foreign_key :channel_works, :users
    add_foreign_key :channel_works, :works
    add_foreign_key :channel_works, :channels
  end
end
