class CreateDbActivities < ActiveRecord::Migration
  def change
    create_table :db_activities do |t|
      t.integer :user_id, null: false
      t.integer :recipient_id
      t.string :recipient_type
      t.integer :trackable_id, null: false
      t.string :trackable_type, null: false
      t.string :action, null: false
      t.timestamps null: false
    end

    add_index :db_activities, [:recipient_id, :recipient_type]
    add_index :db_activities, [:trackable_id, :trackable_type]

    add_foreign_key :db_activities, :users
  end
end
