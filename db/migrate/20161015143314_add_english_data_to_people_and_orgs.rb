# frozen_string_literal: true

class AddEnglishDataToPeopleAndOrgs < ActiveRecord::Migration[5.0]
  def change
    add_column :people, :name_en, :string, null: false, default: ""
    add_column :people, :nickname_en, :string, null: false, default: ""
    add_column :people, :gender_en, :string, null: false, default: ""
    add_column :people, :url_en, :string, null: false, default: ""
    add_column :people, :wikipedia_url_en, :string, null: false, default: ""
    add_column :people, :twitter_username_en, :string, null: false, default: ""
    add_column :people, :blood_type_en, :string, null: false, default: ""

    add_column :organizations, :name_en, :string, null: false, default: ""
    add_column :organizations, :url_en, :string, null: false, default: ""
    add_column :organizations, :wikipedia_url_en, :string, null: false, default: ""
    add_column :organizations, :twitter_username_en, :string, null: false, default: ""
  end
end
