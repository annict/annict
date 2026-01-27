# typed: false
# frozen_string_literal: true

class Provider < ApplicationRecord
  extend Enumerize

  enumerize :name, in: %i[gumroad twitter]

  belongs_to :user

  def token_secret=(secret)
    value = secret if name == "twitter"
    write_attribute(:token_secret, value)
  end
end
