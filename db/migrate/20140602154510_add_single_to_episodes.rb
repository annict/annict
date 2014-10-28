class AddSingleToEpisodes < ActiveRecord::Migration
  def change
    add_column :episodes, :single, :boolean, default: false, after: :title
  end
end
