class CreateSeasons < ActiveRecord::Migration[4.2]
  def change
    create_table :seasons do |t|
      t.string :name, null: false
      t.string :slug, null: false
      t.timestamps
    end

    add_index :seasons, :slug, unique: true
  end
end
