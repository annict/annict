# frozen_string_literal: true

class AddEnglishDataToWorks < ActiveRecord::Migration[5.0]
  def change
    add_column :works, :title_ro, :string, default: "", null: false
    add_column :works, :title_en, :string, default: "", null: false
    add_column :works, :official_site_en_url, :string, default: "", null: false
    add_column :works, :wikipedia_en_url, :string, default: "", null: false
    add_column :works, :synopsis, :text, default: "", null: false
    add_column :works, :synopsis_en, :text, default: "", null: false
    add_column :works, :synopsis_source, :string, default: "", null: false
    add_column :works, :synopsis_en_source, :string, default: "", null: false
    add_column :works, :mal_anime_id, :integer
  end
end
