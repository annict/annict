class AddScTidToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :sc_tid, :integer, after: :season_id
    add_index  :works, :sc_tid, unique: true
  end
end
