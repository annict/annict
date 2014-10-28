class AddScLastUpdateToPrograms < ActiveRecord::Migration
  def change
    add_column :programs, :sc_last_update, :datetime, after: :started_at
  end
end
