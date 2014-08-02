class AddNicochStartedAtToWorks < ActiveRecord::Migration
  def change
    add_column :works, :nicoch_started_at, :datetime, after: :released_at
  end
end
