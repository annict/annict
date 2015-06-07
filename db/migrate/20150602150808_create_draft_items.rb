class CreateDraftItems < ActiveRecord::Migration
  def change
    create_table :draft_items do |t|
      t.integer :item_id
      t.integer :work_id, null: false
      t.string :name, null: false
      t.string :url, null: false
      t.boolean :main, null: false, default: false
      t.string :tombo_image_file_name, null: false
      t.string :tombo_image_content_type, null: false
      t.integer :tombo_image_file_size, null: false
      t.datetime :tombo_image_updated_at, null: false
      t.timestamps null: false
    end

    add_index :draft_items, :item_id
    add_index :draft_items, :work_id
    add_foreign_key :draft_items, :items
    add_foreign_key :draft_items, :works
  end
end
