class CreateChannelWorks < ActiveRecord::Migration
  def change
    create_table :channel_works do |t|
      t.integer :user_id,    null: false
      t.integer :work_id,    null: false
      t.integer :channel_id, null: false
      t.timestamps

      t.foreign_key :users,    dependent: :delete
      t.foreign_key :works,    dependent: :delete
      t.foreign_key :channels, dependent: :delete
    end

    add_index :channel_works, [:user_id, :work_id]
    add_index :channel_works, [:user_id, :work_id, :channel_id], unique: true
  end
end
