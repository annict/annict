class AddScCountToEpisodes < ActiveRecord::Migration
  def change
    add_column :episodes, :sc_count, :integer, after: :sort_number
    add_index  :episodes, [:work_id, :sc_count], unique: true
  end
end
