class CreateDraftOrganizations < ActiveRecord::Migration
  def change
    create_table :draft_organizations do |t|
      t.integer :organization_id
      t.string :name, null: false
      t.string :url
      t.string :wikipedia_url
      t.string :twitter_username
      t.timestamps null: false
    end

    add_index :draft_organizations, :organization_id
    add_index :draft_organizations, :name

    add_foreign_key :draft_organizations, :organizations
  end
end
