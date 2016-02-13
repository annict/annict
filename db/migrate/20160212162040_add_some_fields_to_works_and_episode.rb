class AddSomeFieldsToWorksAndEpisode < ActiveRecord::Migration
  def change
    create_table :number_formats do |t|
      t.string :name, null: false
      t.string :data, null: false, array: true, default: []
      t.integer :sort_number, null: false, default: 0
      t.timestamps null: false
    end

    add_index :number_formats, :name, unique: true

    add_column :works, :number_format_id, :integer
    add_index :works, :number_format_id
    add_foreign_key :works, :number_formats

    add_column :draft_works, :number_format_id, :integer
    add_index :draft_works, :number_format_id
    add_foreign_key :draft_works, :number_formats

    add_column :episodes, :raw_number, :string

    add_column :draft_episodes, :raw_number, :string
  end
end
