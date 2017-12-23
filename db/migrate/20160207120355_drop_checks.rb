class DropChecks < ActiveRecord::Migration[4.2]
  def change
    drop_table :checks
  end
end
