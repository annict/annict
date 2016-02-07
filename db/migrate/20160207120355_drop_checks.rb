class DropChecks < ActiveRecord::Migration
  def change
    drop_table :checks
  end
end
