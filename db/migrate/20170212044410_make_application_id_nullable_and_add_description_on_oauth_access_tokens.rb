# frozen_string_literal: true

class MakeApplicationIdNullableAndAddDescriptionOnOauthAccessTokens < ActiveRecord::Migration[5.0]
  def change
    add_column :oauth_access_tokens, :description, :string, null: false, default: ""
    change_column_null :oauth_access_tokens, :application_id, true
  end
end
