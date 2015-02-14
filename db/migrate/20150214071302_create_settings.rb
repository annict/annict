class CreateSettings < ActiveRecord::Migration
  def change
    create_table :settings do |t|
      t.integer :user_id, null: false
      t.boolean :hide_checkin_comment, null: false, default: true
      t.timestamps null: false
    end

    add_index :settings, :user_id
    add_foreign_key :settings, :users
  end
end
