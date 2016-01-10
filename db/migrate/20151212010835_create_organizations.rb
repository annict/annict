class CreateOrganizations < ActiveRecord::Migration
  def change
    create_table :organizations do |t|
      t.string :name, null: false
      t.string :url
      t.string :wikipedia_url
      t.string :twitter_username
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :organizations, :name, unique: true
    add_index :organizations, :aasm_state
  end
end
