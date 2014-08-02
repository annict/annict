class CreateReceptions < ActiveRecord::Migration
  def change
    create_table :receptions do |t|
      t.integer :user_id,    null: false
      t.integer :channel_id, null: false
      t.timestamps

      t.foreign_key :users,    dependent: :delete
      t.foreign_key :channels, dependent: :delete
    end

    add_index :receptions, [:user_id, :channel_id], unique: true
  end
end
