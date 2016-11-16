# frozen_string_literal: true

class CreateCharacterImages < ActiveRecord::Migration[5.0]
  def change
    create_table :character_images do |t|
      t.integer :character_id, null: false
      t.integer :user_id, null: false
      t.string :attachment_file_name, null: false
      t.integer :attachment_file_size, null: false
      t.string :attachment_content_type, null: false
      t.datetime :attachment_updated_at, null: false
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :character_images, :character_id
    add_index :character_images, :user_id
    add_index :character_images, :aasm_state

    add_foreign_key :character_images, :characters
    add_foreign_key :character_images, :users
  end
end
