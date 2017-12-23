class AddTwitterFieldsToWorks < ActiveRecord::Migration[4.2]
  def change
    add_column :works, :twitter_username, :string
    add_column :works, :twitter_hashtag, :string
  end
end
