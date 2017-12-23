class CreateDraftPeople < ActiveRecord::Migration[4.2]
  def change
    create_table :draft_people do |t|
      t.integer :person_id
      t.integer :prefecture_id
      t.string :name, null: false
      t.string :name_kana
      t.string :nickname
      t.string :gender
      t.string :url
      t.string :wikipedia_url
      t.string :twitter_username
      t.date :birthday
      t.string :blood_type
      t.integer :height
      t.timestamps null: false
    end

    add_index :draft_people, :person_id
    add_index :draft_people, :prefecture_id
    add_index :draft_people, :name

    add_foreign_key :draft_people, :people
    add_foreign_key :draft_people, :prefectures
  end
end
