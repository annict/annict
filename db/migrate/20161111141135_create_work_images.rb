# frozen_string_literal: true

class CreateWorkImages < ActiveRecord::Migration[5.0]
  def change
    create_table :work_images do |t|
      t.integer :work_id, null: false
      t.integer :user_id, null: false
      t.string :attachment_file_name, null: false
      t.integer :attachment_file_size, null: false
      t.string :attachment_content_type, null: false
      t.datetime :attachment_updated_at, null: false
      t.string :copyright, null: false, default: ""
      t.string :asin, null: false, default: ""
      t.timestamps null: false
    end

    add_index :work_images, :work_id
    add_index :work_images, :user_id
    add_index :work_images, :aasm_state

    add_foreign_key :work_images, :works
    add_foreign_key :work_images, :users
  end
end
