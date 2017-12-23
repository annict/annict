class CreateEditRequestComments < ActiveRecord::Migration[4.2]
  def change
    create_table :edit_request_comments do |t|
      t.integer :edit_request_id, null: false
      t.integer :user_id, null: false
      t.text :body, null: false
      t.timestamps null: false
    end

    add_index :edit_request_comments, :edit_request_id
    add_index :edit_request_comments, :user_id
    add_foreign_key :edit_request_comments, :edit_requests, on_delete: :cascade
    add_foreign_key :edit_request_comments, :users, on_delete: :cascade
  end
end
