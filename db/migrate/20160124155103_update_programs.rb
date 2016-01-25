class UpdatePrograms < ActiveRecord::Migration
  def change
    add_column :programs, :sc_pid, :integer
    add_column :programs, :rebroadcast, :boolean, null: false, default: false

    add_index :programs, :sc_pid, unique: true
    remove_index :programs, name: :programs_channel_id_episode_id_key
  end
end
