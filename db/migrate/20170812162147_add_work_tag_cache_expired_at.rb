# frozen_string_literal: true

class AddWorkTagCacheExpiredAt < ActiveRecord::Migration[5.1]
  def change
    add_column :users, :work_tag_cache_expired_at, :datetime
    add_column :users, :work_comment_cache_expired_at, :datetime
  end
end
