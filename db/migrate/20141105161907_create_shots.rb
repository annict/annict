class CreateShots < ActiveRecord::Migration[4.2]
  def change
    create_table :shots do |t|
      t.integer :user_id, null: false
      t.string :image_uid, null: false
      t.timestamps
    end

    add_index :shots, :image_uid, unique: true

    add_foreign_key :shots, :users
  end
end
