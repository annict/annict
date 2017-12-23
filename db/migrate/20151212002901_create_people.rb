class CreatePeople < ActiveRecord::Migration[4.2]
  def change
    create_table :people do |t|
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
      t.string :aasm_state, null: false, default: "published"
      t.timestamps null: false
    end

    add_index :people, :prefecture_id
    add_index :people, :name, unique: true
    add_index :people, :aasm_state

    add_foreign_key :people, :prefectures
  end
end
