class CreateOrganizations < ActiveRecord::Migration
  def change
    create_table :organizations do |t|
      t.string :name, null: false
      t.string :url
      t.string :wikipedia_url
      t.timestamps null: false
    end

    add_index :organizations, :name, unique: true
  end
end
