# frozen_string_literal: true

class AddPrivacyPolicyAgreedToSettings < ActiveRecord::Migration[5.2]
  def change
    add_column :settings, :privacy_policy_agreed, :boolean, default: false, null: false
  end
end
