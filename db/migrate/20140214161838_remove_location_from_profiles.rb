class RemoveLocationFromProfiles < ActiveRecord::Migration[4.2]
  def change
    remove_column :profiles, :location
  end
end
