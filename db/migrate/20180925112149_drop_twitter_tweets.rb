# frozen_string_literal: true

class DropTwitterTweets < ActiveRecord::Migration[5.2]
  def change
    drop_table :twitter_tweets
    drop_table :twitter_users
  end
end
