class RemoveUrlFromProfiles < ActiveRecord::Migration
  def change
    remove_column :profiles, :url
  end
end
