# frozen_string_literal: true

class AddCommentableToComments < ActiveRecord::Migration[6.1]
  def change
    add_column :comments, :commentable_id, :bigint
    add_column :comments, :commentable_type, :string
    add_index :comments, %i[commentable_id commentable_type]
  end
end
