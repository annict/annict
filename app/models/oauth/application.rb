# typed: false
# frozen_string_literal: true

class Oauth::Application < ApplicationRecord
  include ::Doorkeeper::Orm::ActiveRecord::Mixins::Application

  include BatchDestroyable

  scope :available, -> { where(deleted_at: nil).where.not(owner: nil) }
  scope :unavailable, -> {
    unscoped.where.not(deleted_at: nil).or(where(owner: nil))
  }
  scope :authorized, -> { where(oauth_access_tokens: {revoked_at: nil}) }
end
