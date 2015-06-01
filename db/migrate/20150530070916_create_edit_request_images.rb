class CreateEditRequestImages < ActiveRecord::Migration
  def change
    create_table :edit_request_images do |t|
      t.integer :edit_request_id, null: false
      t.string :image_file_name, null: false
      t.string :image_content_type, null: false
      t.integer :image_file_size, null: false
      t.datetime :image_updated_at, null: false
      t.timestamps null: false
    end

    add_index :edit_request_images, :edit_request_id
    add_foreign_key :edit_request_images, :edit_requests, on_delete: :cascade
  end
end
