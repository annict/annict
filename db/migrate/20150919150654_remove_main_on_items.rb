class RemoveMainOnItems < ActiveRecord::Migration
  def change
    remove_column :items, :main
  end
end
