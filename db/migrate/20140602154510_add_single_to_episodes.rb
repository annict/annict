class AddSingleToEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :episodes, :single, :boolean, default: false, after: :title
  end
end
