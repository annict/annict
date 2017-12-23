class CreateDraftCasts < ActiveRecord::Migration[4.2]
  def change
    create_table :draft_casts do |t|
      t.integer :cast_id
      t.integer :person_id, null: false
      t.integer :work_id, null: false
      t.string :name, null: false
      t.string :part, null: false
      t.integer :sort_number, null: false, default: 0
      t.timestamps null: false
    end

    add_index :draft_casts, :cast_id
    add_index :draft_casts, :person_id
    add_index :draft_casts, :work_id
    add_index :draft_casts, :sort_number

    add_foreign_key :draft_casts, :casts
    add_foreign_key :draft_casts, :people
    add_foreign_key :draft_casts, :works
  end
end
