class AddScNumberToEpisodes < ActiveRecord::Migration
  def change
    add_column :episodes, :sc_number, :string, after: :sort_number
    add_index  :episodes, :sc_number
    add_index  :episodes, [:work_id, :sc_number], unique: true
  end
end
