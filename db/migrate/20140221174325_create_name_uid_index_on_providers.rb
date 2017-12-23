class CreateNameUidIndexOnProviders < ActiveRecord::Migration[4.2]
  def change
    add_index :providers, [:name, :uid], unique: true
  end
end
