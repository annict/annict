class CreateLikes < ActiveRecord::Migration
  def change
    create_table :likes do |t|
      t.integer :user_id,        null: false
      t.integer :recipient_id,   null: false
      t.string  :recipient_type, null: false
      t.timestamps
    end

    add_column :checkins, :likes_count, :integer, null: false, default: 0, after: :comments_count
    add_column :comments, :likes_count, :integer, null: false, default: 0, after: :body

    add_index :likes, [:recipient_id, :recipient_type]

    add_foreign_key :likes, :users
  end
end
