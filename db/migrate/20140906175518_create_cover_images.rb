class CreateCoverImages < ActiveRecord::Migration[4.2]
  def change
    create_table :cover_images do |t|
      t.integer :work_id, null: false
      t.string :file_name, null: false
      t.string :location, null: false
      t.timestamps
    end

    add_foreign_key :cover_images, :works
  end
end
