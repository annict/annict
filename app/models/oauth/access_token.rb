# frozen_string_literal: true

class Oauth::AccessToken < ApplicationRecord
  include ::Doorkeeper::Orm::ActiveRecord::Mixins::AccessToken

  include BatchDestroyable

  belongs_to :owner, class_name: "User", foreign_key: :resource_owner_id

  scope :available, -> { where(revoked_at: nil) }
  scope :personal, -> { where(application_id: nil) }

  validates :description, presence: {on: :personal}

  before_validation :generate_token, on: %i[create personal]

  def writable?
    scopes.include?("write")
  end
end
