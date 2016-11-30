# frozen_string_literal: true

class CreateCharacters < ActiveRecord::Migration[5.0]
  def change
    create_table :characters do |t|
      t.string :name, null: false
      t.string :name_kana, null: false, default: ""
      t.string :name_en, null: false, default: ""
      t.string :kind, null: false, default: ""
      t.string :kind_en, null: false, default: ""
      t.string :nickname, null: false, default: ""
      t.string :nickname_en, null: false, default: ""
      t.string :birthday, null: false, default: ""
      t.string :birthday_en, null: false, default: ""
      t.string :age, null: false, default: ""
      t.string :age_en, null: false, default: ""
      t.string :blood_type, null: false, default: ""
      t.string :blood_type_en, null: false, default: ""
      t.string :height, null: false, default: ""
      t.string :height_en, null: false, default: ""
      t.string :weight, null: false, default: ""
      t.string :weight_en, null: false, default: ""
      t.string :nationality, null: false, default: ""
      t.string :nationality_en, null: false, default: ""
      t.string :occupation, null: false, default: ""
      t.string :occupation_en, null: false, default: ""
      t.text :description, null: false, default: ""
      t.text :description_en, null: false, default: ""
      t.string :aasm_state, null: false, default: "published"
      t.timestamps
    end

    add_index :characters, [:name, :kind], unique: true
  end
end
