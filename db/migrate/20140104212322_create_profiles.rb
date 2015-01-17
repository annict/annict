class CreateProfiles < ActiveRecord::Migration
  def change
    create_table :profiles do |t|
      t.integer     :user_id,     null: false
      t.string      :name,        null: false, default: ''
      t.string      :location,    null: false, default: ''
      t.string      :url,         null: false, default: ''
      t.string      :description, null: false, default: ''
      t.timestamps
    end

    add_index       :profiles, :user_id, unique: true
    add_foreign_key :profiles, :users
  end
end
