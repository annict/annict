class CreateComments < ActiveRecord::Migration
  def change
    create_table :comments do |t|
      t.integer :user_id,    null: false
      t.integer :checkin_id, null: false
      t.text    :body,       null: false
      t.timestamps

      t.foreign_key :users,    dependent: :delete
      t.foreign_key :checkins, dependent: :delete
    end
  end
end
