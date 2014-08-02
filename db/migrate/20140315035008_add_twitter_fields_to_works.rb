class AddTwitterFieldsToWorks < ActiveRecord::Migration
  def change
    add_column :works, :twitter_username, :string
    add_column :works, :twitter_hashtag, :string
  end
end
