class AddReleasedAtAboutToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :released_at_about, :string
  end
end
