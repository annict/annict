class AddUrlToProfiles < ActiveRecord::Migration[4.2]
  def change
    add_column :profiles, :url, :string
  end
end
