# frozen_string_literal: true

class AddEnglishDataToEpisodes < ActiveRecord::Migration[5.0]
  def change
    add_column :episodes, :title_ro, :string, default: "", null: false
    add_column :episodes, :title_en, :string, default: "", null: false
  end
end
