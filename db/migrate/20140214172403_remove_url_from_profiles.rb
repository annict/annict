class RemoveUrlFromProfiles < ActiveRecord::Migration[4.2]
  def change
    remove_column :profiles, :url
  end
end
