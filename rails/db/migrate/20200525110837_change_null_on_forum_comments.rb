# frozen_string_literal: true

class ChangeNullOnForumComments < ActiveRecord::Migration[6.0]
  def change
    change_column_null :forum_comments, :user_id, true
  end
end
