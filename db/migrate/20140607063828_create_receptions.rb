class CreateReceptions < ActiveRecord::Migration
  def change
    create_table :receptions do |t|
      t.integer :user_id,    null: false
      t.integer :channel_id, null: false
      t.timestamps
    end

    add_index :receptions, [:user_id, :channel_id], unique: true

    add_foreign_key :receptions, :users
    add_foreign_key :receptions, :channels
  end
end
