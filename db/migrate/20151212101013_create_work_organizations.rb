class CreateWorkOrganizations < ActiveRecord::Migration
  def change
    create_table :work_organizations do |t|
      t.integer :work_id, null: false
      t.integer :organization_id, null: false
      t.string :role, null: false
      t.string :role_other
      t.timestamps null: false
    end

    add_index :work_organizations, :work_id
    add_index :work_organizations, :organization_id
    add_index :work_organizations, [:work_id, :organization_id], unique: true

    add_foreign_key :work_organizations, :works
    add_foreign_key :work_organizations, :organizations
  end
end
