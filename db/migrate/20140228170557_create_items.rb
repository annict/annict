class CreateItems < ActiveRecord::Migration
  def change
    create_table :items do |t|
      t.integer :work_id
      t.string  :name,      null: false
      t.string  :url,       null: false
      t.string  :image_uid, null: false
      t.boolean :main,      null: false, default: false
      t.timestamps
    end

    add_foreign_key :items, :works
  end
end
