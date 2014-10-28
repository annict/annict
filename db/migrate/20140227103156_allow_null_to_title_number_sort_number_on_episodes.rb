class AllowNullToTitleNumberSortNumberOnEpisodes < ActiveRecord::Migration
  def change
    change_column :episodes, :number, :string, null: true
    change_column :episodes, :title,  :string, null: true, default: nil
  end
end
