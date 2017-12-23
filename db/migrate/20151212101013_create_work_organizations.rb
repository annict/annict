class CreateWorkOrganizations < ActiveRecord::Migration[4.2]
  def change
    create_table :work_organizations do |t|
      t.integer :work_id, null: false
      t.integer :organization_id, null: false
      t.string :role, null: false
      t.string :role_other
      t.string :aasm_state, null: false, default: "published"
      t.integer :sort_number, null: false, default: 0
      t.timestamps null: false
    end

    add_index :work_organizations, :work_id
    add_index :work_organizations, :organization_id
    add_index :work_organizations, [:work_id, :organization_id], unique: true
    add_index :work_organizations, :aasm_state
    add_index :work_organizations, :sort_number

    add_foreign_key :work_organizations, :works
    add_foreign_key :work_organizations, :organizations
  end
end
