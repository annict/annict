# frozen_string_literal: true

class ChangeResourceOwnerIdToUserIdAndAddGuestIdOnOauthAccessTokens < ActiveRecord::Migration[5.2]
  def change
    change_column_null :oauth_access_tokens, :resource_owner_id, true

    add_column :oauth_access_tokens, :guest_id, :integer
    add_index :oauth_access_tokens, :guest_id
    add_foreign_key :oauth_access_tokens, :guests
  end
end
