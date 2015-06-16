class CreateEditRequestParticipants < ActiveRecord::Migration
  def change
    create_table :edit_request_participants do |t|
      t.integer :edit_request_id, null: false
      t.integer :user_id, null: false
      t.timestamps null: false
    end

    add_index :edit_request_participants, :edit_request_id
    add_index :edit_request_participants, :user_id
    add_index :edit_request_participants, [:edit_request_id, :user_id], unique: true

    add_foreign_key :edit_request_participants, :edit_requests
    add_foreign_key :edit_request_participants, :users
  end
end
