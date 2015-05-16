class CreateEditRequests < ActiveRecord::Migration
  def change
    create_table :edit_requests do |t|
      t.integer :user_id, null: false
      t.integer :kind, null: false
      t.integer :status, null: false, default: 1
      t.integer :resource_id
      t.string :resource_type
      t.integer :trackable_id
      t.string :trackable_type
      t.json :draft_resource_params, null: false
      t.string :title, null: false
      t.text :body
      t.datetime :merged_at
      t.datetime :closed_at
      t.timestamps null: false
    end

    add_index :edit_requests, :user_id
    add_index :edit_requests, [:resource_id, :resource_type]
    add_index :edit_requests, [:trackable_id, :trackable_type]
    add_foreign_key :edit_requests, :users, on_delete: :cascade
  end
end
