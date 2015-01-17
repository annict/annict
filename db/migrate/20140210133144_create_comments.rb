class CreateComments < ActiveRecord::Migration
  def change
    create_table :comments do |t|
      t.integer :user_id,    null: false
      t.integer :checkin_id, null: false
      t.text    :body,       null: false
      t.timestamps
    end

    add_foreign_key :comments, :users
    add_foreign_key :comments, :checkins
  end
end
