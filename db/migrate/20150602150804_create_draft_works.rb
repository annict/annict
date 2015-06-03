class CreateDraftWorks < ActiveRecord::Migration
  def change
    create_table :draft_works do |t|
      t.integer :work_id
      t.integer :season_id
      t.integer :sc_tid
      t.string :title, null: false
      t.integer :media, null: false
      t.string :official_site_url, default: "", null: false
      t.string :wikipedia_url, default: "", null: false
      t.date :released_at
      t.string :twitter_username
      t.string :twitter_hashtag
      t.string :released_at_about
      t.timestamps null: false
    end

    add_index :draft_works, :sc_tid, unique: true
    add_index :draft_works, :work_id
    add_index :draft_works, :season_id
    add_foreign_key :draft_works, :works
    add_foreign_key :draft_works, :seasons
  end
end
