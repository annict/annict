class CreateLists < ActiveRecord::Migration[4.2]
  def change
    create_table :statuses do |t|
      t.integer     :user_id, null: false
      t.integer     :work_id, null: false
      t.integer     :kind,    null: false
      t.timestamps
    end

    add_foreign_key :statuses, :users
    add_foreign_key :statuses, :works
  end
end
