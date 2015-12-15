class CreateCastParticipations < ActiveRecord::Migration
  def change
    create_table :cast_participations do |t|
      t.integer :person_id, null: false
      t.integer :work_id, null: false
      t.string :name
      t.string :character_name, null: false
      t.timestamps null: false
    end

    add_index :cast_participations, :person_id
    add_index :cast_participations, :work_id

    add_foreign_key :cast_participations, :people
    add_foreign_key :cast_participations, :works
  end
end
