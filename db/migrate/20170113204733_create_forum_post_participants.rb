# frozen_string_literal: true

class CreateForumPostParticipants < ActiveRecord::Migration[5.0]
  def change
    create_table :forum_post_participants do |t|
      t.integer :forum_post_id, null: false
      t.integer :user_id, null: false
      t.timestamps null: false
    end

    add_index :forum_post_participants, :forum_post_id
    add_index :forum_post_participants, :user_id
    add_index :forum_post_participants, %i[forum_post_id user_id], unique: true

    add_foreign_key :forum_post_participants, :forum_posts
    add_foreign_key :forum_post_participants, :users
  end
end
