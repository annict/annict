class CreateCoverImages < ActiveRecord::Migration
  def change
    create_table :cover_images do |t|
      t.integer :work_id, null: false
      t.string :file_name, null: false
      t.string :location, null: false
      t.timestamps

      t.foreign_key :works, dependent: :delete
    end
  end
end
