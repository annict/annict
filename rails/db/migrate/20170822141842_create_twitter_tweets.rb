# frozen_string_literal: true

class CreateTwitterTweets < ActiveRecord::Migration[5.1]
  def change
    create_table :twitter_tweets do |t|
      t.integer :twitter_user_id, null: false
      t.string :user_screen_name, null: false
      t.string :user_name, null: false
      t.string :tweet_id, null: false
      t.text :text, null: false
      t.timestamps null: false
    end

    add_index :twitter_tweets, :twitter_user_id
    add_index :twitter_tweets, :tweet_id, unique: true

    add_foreign_key :twitter_tweets, :twitter_users
  end
end
