class CreateDraftWorkOrganizations < ActiveRecord::Migration
  def change
    create_table :draft_work_organizations do |t|
      t.integer :work_organization_id
      t.integer :work_id, null: false
      t.integer :organization_id, null: false
      t.string :role, null: false
      t.string :role_other
      t.integer :sort_number, null: false, default: 0
      t.timestamps null: false
    end

    add_index :draft_work_organizations, :work_organization_id
    add_index :draft_work_organizations, :work_id
    add_index :draft_work_organizations, :organization_id
    add_index :draft_work_organizations, :sort_number

    add_foreign_key :draft_work_organizations, :work_organizations
    add_foreign_key :draft_work_organizations, :works
    add_foreign_key :draft_work_organizations, :organizations
  end
end
