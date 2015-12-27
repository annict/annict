class CreateCasts < ActiveRecord::Migration
  def change
    create_table :casts do |t|
      t.integer :person_id, null: false
      t.integer :work_id, null: false
      t.string :name, null: false
      t.string :part, null: false
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :casts, :person_id
    add_index :casts, :work_id
    add_index :casts, :aasm_state

    add_foreign_key :casts, :people
    add_foreign_key :casts, :works
  end
end
