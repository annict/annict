# typed: false
# frozen_string_literal: true

class Provider < ApplicationRecord
  extend Enumerize

  enumerize :name, in: %i[gumroad twitter]

  belongs_to :user

  scope :token_available, -> {
    where(token_expires_at: nil)
      .or(where("token_expires_at > ?", Time.now.to_i))
  }

  def token_secret=(secret)
    value = secret if name == "twitter"
    write_attribute(:token_secret, value)
  end
end
