# frozen_string_literal: true

class CreateReviewComments < ActiveRecord::Migration[5.1]
  def change
    create_table :review_comments do |t|
      t.integer :user_id, null: false
      t.integer :review_id, null: false
      t.integer :work_id, null: false
      t.text :body, null: false
      t.timestamps null: false
    end

    add_index :review_comments, :user_id
    add_index :review_comments, :review_id
    add_index :review_comments, :work_id

    add_foreign_key :review_comments, :users
    add_foreign_key :review_comments, :reviews
    add_foreign_key :review_comments, :works
  end
end
