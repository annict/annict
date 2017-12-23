class AddScCountToEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :episodes, :sc_count, :integer, after: :sort_number
    add_index  :episodes, [:work_id, :sc_count], unique: true
  end
end
