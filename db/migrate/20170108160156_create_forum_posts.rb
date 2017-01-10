# frozen_string_literal: true

class CreateForumPosts < ActiveRecord::Migration[5.0]
  def change
    create_table :forum_posts do |t|
      t.integer :user_id, null: false
      t.integer :forum_category_id, null: false
      t.string :title, null: false
      t.text :body, null: false, default: ""
      t.integer :forum_comments_count, null: false, default: 0
      t.datetime :edited_at,
        comment: "The datetime which user has changed title, body and so on."
      t.datetime :last_commented_at, null: false
      t.timestamps null: false
    end

    add_index :forum_posts, :user_id
    add_index :forum_posts, :forum_category_id

    add_foreign_key :forum_posts, :users
    add_foreign_key :forum_posts, :forum_categories
  end
end
