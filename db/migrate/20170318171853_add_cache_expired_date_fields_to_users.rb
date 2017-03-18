# frozen_string_literal: true

class AddCacheExpiredDateFieldsToUsers < ActiveRecord::Migration[5.0]
  def change
    add_column :users, :record_cache_expired_at, :datetime
    add_column :users, :status_cache_expired_at, :datetime
  end
end
