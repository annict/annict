class AddScLastUpdateToPrograms < ActiveRecord::Migration[4.2]
  def change
    add_column :programs, :sc_last_update, :datetime, after: :started_at
  end
end
