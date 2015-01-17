class AddReleasedAtAboutToWorks < ActiveRecord::Migration
  def change
    add_column :works, :released_at_about, :string
  end
end
