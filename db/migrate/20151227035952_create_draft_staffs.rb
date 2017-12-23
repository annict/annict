class CreateDraftStaffs < ActiveRecord::Migration[4.2]
  def change
    create_table :draft_staffs do |t|
      t.integer :staff_id
      t.integer :person_id, null: false
      t.integer :work_id, null: false
      t.string :name, null: false
      t.string :role, null: false
      t.string :role_other
      t.integer :sort_number, null: false, default: 0
      t.timestamps null: false
    end

    add_index :draft_staffs, :staff_id
    add_index :draft_staffs, :person_id
    add_index :draft_staffs, :work_id
    add_index :draft_staffs, :sort_number

    add_foreign_key :draft_staffs, :staffs
    add_foreign_key :draft_staffs, :people
    add_foreign_key :draft_staffs, :works
  end
end
