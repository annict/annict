class CreateStaffParticipations < ActiveRecord::Migration
  def change
    create_table :staff_participations do |t|
      t.integer :person_id, null: false
      t.integer :work_id, null: false
      t.string :name
      t.string :role, null: false
      t.string :role_other
      t.timestamps null: false
    end

    add_index :staff_participations, :person_id
    add_index :staff_participations, :work_id

    add_foreign_key :staff_participations, :people
    add_foreign_key :staff_participations, :works
  end
end
