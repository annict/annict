class CreateStaffs < ActiveRecord::Migration
  def change
    create_table :staffs do |t|
      t.integer :person_id, null: false
      t.integer :work_id, null: false
      t.string :name, null: false
      t.string :role, null: false
      t.string :role_other
      t.timestamps null: false
    end

    add_index :staffs, :person_id
    add_index :staffs, :work_id

    add_foreign_key :staffs, :people
    add_foreign_key :staffs, :works
  end
end
