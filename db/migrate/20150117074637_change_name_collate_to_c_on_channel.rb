class ChangeNameCollateToCOnChannel < ActiveRecord::Migration[4.2]
  def up
    sql = 'ALTER TABLE channels ALTER COLUMN name TYPE varchar COLLATE "C"'
    ActiveRecord::Base.connection.execute(sql)
  end

  def down
    ActiveRecord::IrreversibleMigration
  end
end
