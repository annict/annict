class CreateFinishedTips < ActiveRecord::Migration
  def change
    create_table :finished_tips do |t|
      t.integer :user_id, null: false
      t.integer :tip_id, null: false
      t.timestamps null: false
    end

    add_index :finished_tips, [:user_id, :tip_id], unique: true

    add_foreign_key :finished_tips, :users
    add_foreign_key :finished_tips, :tips
  end
end
