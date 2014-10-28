class CreateNameUidIndexOnProviders < ActiveRecord::Migration
  def change
    add_index :providers, [:name, :uid], unique: true
  end
end
