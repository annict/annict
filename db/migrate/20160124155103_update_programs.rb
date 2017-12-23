class UpdatePrograms < ActiveRecord::Migration[4.2]
  def change
    add_column :programs, :sc_pid, :integer
    add_column :programs, :rebroadcast, :boolean, null: false, default: false

    add_index :programs, :sc_pid, unique: true

    add_column :draft_programs, :rebroadcast, :boolean, null: false, default: false
  end
end
