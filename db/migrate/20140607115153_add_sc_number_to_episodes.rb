class AddScNumberToEpisodes < ActiveRecord::Migration[4.2]
  def change
    add_column :episodes, :sc_number, :string, after: :sort_number
    add_index  :episodes, :sc_number
    add_index  :episodes, [:work_id, :sc_number], unique: true
  end
end
