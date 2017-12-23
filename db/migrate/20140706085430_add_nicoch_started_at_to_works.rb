class AddNicochStartedAtToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :nicoch_started_at, :datetime, after: :released_at
  end
end
