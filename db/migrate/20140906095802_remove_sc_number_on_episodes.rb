class RemoveScNumberOnEpisodes < ActiveRecord::Migration[4.2]
  def change
    remove_index :episodes, name: 'index_episodes_on_work_id_and_sc_number'
    remove_index :episodes, :sc_number
    remove_column :episodes, :sc_number
  end
end
