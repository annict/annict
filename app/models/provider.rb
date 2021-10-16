# frozen_string_literal: true

# == Schema Information
#
# Table name: providers
#
#  id               :bigint           not null, primary key
#  name             :string(510)      not null
#  token            :string(510)      not null
#  token_expires_at :integer
#  token_secret     :string(510)
#  uid              :string(510)      not null
#  created_at       :datetime
#  updated_at       :datetime
#  user_id          :bigint           not null
#
# Indexes
#
#  providers_name_uid_key  (name,uid) UNIQUE
#  providers_user_id_idx   (user_id)
#
# Foreign Keys
#
#  providers_user_id_fk  (user_id => users.id) ON DELETE => cascade
#

class Provider < ApplicationRecord
  extend Enumerize

  enumerize :name, in: %i[facebook gumroad twitter]

  belongs_to :user

  scope :token_available, -> {
    where(token_expires_at: nil)
      .or(where("token_expires_at > ?", Time.now.to_i))
  }

  def token_expires_at=(expires_at)
    value = expires_at if name == "facebook"
    write_attribute(:token_expires_at, value)
  end

  def token_secret=(secret)
    value = secret if name == "twitter"
    write_attribute(:token_secret, value)
  end
end
