class RemoveNicochStartedAtOnWorks < ActiveRecord::Migration
  def change
    remove_column :works, :nicoch_started_at
  end
end
