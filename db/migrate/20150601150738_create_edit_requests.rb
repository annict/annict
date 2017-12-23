class CreateEditRequests < ActiveRecord::Migration[4.2]
  def change
    create_table :edit_requests do |t|
      t.integer :user_id, null: false
      t.references :draft_resource, polymorphic: true, null: false
      t.string :title, null: false
      t.text :body
      t.string :aasm_state, null: false, default: "opened"
      t.datetime :published_at
      t.datetime :closed_at
      t.timestamps null: false
    end

    add_index :edit_requests, :user_id
    add_index :edit_requests, [:draft_resource_id, :draft_resource_type],
              name: :index_er_on_drid_and_drtype
    add_foreign_key :edit_requests, :users, on_delete: :cascade
  end
end
