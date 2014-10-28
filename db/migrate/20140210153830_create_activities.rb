class CreateActivities < ActiveRecord::Migration
  def change
    create_table :activities do |t|
      t.integer :user_id,        null: false
      t.integer :recipient_id,   null: false
      t.string  :recipient_type, null: false
      t.integer :trackable_id,   null: false
      t.string  :trackable_type, null: false
      t.string  :action,         null: false
      t.timestamps

      t.foreign_key :users, dependent: :delete
    end

    add_index :activities, [:recipient_id, :recipient_type]
    add_index :activities, [:trackable_id, :trackable_type]
  end
end
