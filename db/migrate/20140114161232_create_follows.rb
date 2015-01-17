class CreateFollows < ActiveRecord::Migration
  def change
    create_table :follows do |t|
      t.integer     :user_id,      null: false
      t.integer     :following_id, null: false
      t.timestamps
    end

    add_index :follows, [:user_id, :following_id], unique: true

    add_foreign_key :follows, :users
    add_foreign_key :follows, :users, column: :following_id
  end
end
