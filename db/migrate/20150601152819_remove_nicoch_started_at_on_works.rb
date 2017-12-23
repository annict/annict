class RemoveNicochStartedAtOnWorks < ActiveRecord::Migration[4.2]
  def change
    remove_column :works, :nicoch_started_at
  end
end
