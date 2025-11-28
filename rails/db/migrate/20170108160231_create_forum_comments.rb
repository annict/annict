# frozen_string_literal: true

class CreateForumComments < ActiveRecord::Migration[5.0]
  def change
    create_table :forum_comments do |t|
      t.integer :user_id, null: false
      t.integer :forum_post_id, null: false
      t.text :body, null: false
      t.datetime :edited_at,
        comment: "The datetime which user has changed body."
      t.timestamps null: false
    end

    add_index :forum_comments, :user_id
    add_index :forum_comments, :forum_post_id

    add_foreign_key :forum_comments, :users
    add_foreign_key :forum_comments, :forum_posts
  end
end
