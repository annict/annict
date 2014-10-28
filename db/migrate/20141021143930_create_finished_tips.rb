class CreateFinishedTips < ActiveRecord::Migration
  def change
    create_table :finished_tips do |t|
      t.integer :user_id, null: false
      t.integer :tip_id, null: false
      t.timestamps null: false

      t.foreign_key :users, dependent: :delete
      t.foreign_key :tips, dependent: :delete
    end
  end
end
