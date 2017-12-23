class RemoveMainOnItems < ActiveRecord::Migration[4.2]
  def change
    remove_column :items, :main
  end
end
