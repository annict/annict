class CreateMemberParticipations < ActiveRecord::Migration
  def change
    create_table :member_participations do |t|
      t.integer :person_id, null: false
      t.integer :organization_id, null: false
      t.timestamps null: false
    end

    add_index :member_participations, :person_id
    add_index :member_participations, :organization_id
    add_index :member_participations, [:person_id, :organization_id], unique: true

    add_foreign_key :member_participations, :people
    add_foreign_key :member_participations, :organizations
  end
end
