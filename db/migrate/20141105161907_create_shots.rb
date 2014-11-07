class CreateShots < ActiveRecord::Migration
  def change
    create_table :shots do |t|
      t.integer :user_id, null: false
      t.string :image_uid, null: false
      t.timestamps

      t.foreign_key :users, dependent: :delete
    end

    add_index :shots, :image_uid, unique: true
  end
end
