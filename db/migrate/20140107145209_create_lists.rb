class CreateLists < ActiveRecord::Migration
  def change
    create_table :statuses do |t|
      t.integer     :user_id, null: false
      t.foreign_key :users,   dependent: :delete
      t.integer     :work_id, null: false
      t.foreign_key :works,   dependent: :delete
      t.integer     :kind,    null: false
      t.timestamps
    end
  end
end
