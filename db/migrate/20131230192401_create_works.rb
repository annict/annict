class CreateWorks < ActiveRecord::Migration[4.2]
  def change
    create_table :works do |t|
      t.string  :title,             null: false
      t.integer :media,             null: false
      t.string  :official_site_url, null: false, default: ''
      t.string  :wikipedia_url,     null: false, default: ''
      t.date    :released_at
      t.timestamps
    end

    add_index :works, :media
    add_index :works, :released_at
  end
end
