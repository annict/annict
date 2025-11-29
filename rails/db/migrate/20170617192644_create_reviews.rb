# frozen_string_literal: true

class CreateReviews < ActiveRecord::Migration[5.1]
  def change
    create_table :reviews do |t|
      t.integer :user_id, null: false
      t.integer :work_id, null: false
      t.string :title, null: false, default: ""
      t.text :body, null: false
      t.string :rating_animation_state
      t.string :rating_music_state
      t.string :rating_story_state
      t.string :rating_character_state
      t.string :rating_overall_state
      t.integer :likes_count, null: false, default: 0
      t.integer :impressions_count, null: false, default: 0
      t.string :aasm_state, null: false, default: "published"
      t.datetime :modified_at
      t.timestamps null: false
    end

    add_index :reviews, :user_id
    add_index :reviews, :work_id

    add_foreign_key :reviews, :users
    add_foreign_key :reviews, :works

    add_column :checkins, :review_id, :integer
    add_index :checkins, :review_id
    add_foreign_key :checkins, :reviews

    add_column :checkins, :aasm_state, :string, null: false, default: "published"

    add_column :works, :no_episodes, :boolean, null: false, default: false

    add_column :settings, :share_review_to_twitter, :boolean, null: false, default: false
    add_column :settings, :share_review_to_facebook, :boolean, null: false, default: false
  end
end
