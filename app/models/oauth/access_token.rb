# frozen_string_literal: true

# == Schema Information
#
# Table name: oauth_access_tokens
#
#  id                     :bigint           not null, primary key
#  description            :string           default(""), not null
#  expires_in             :integer
#  previous_refresh_token :string           default(""), not null
#  refresh_token          :string
#  revoked_at             :datetime
#  scopes                 :string
#  token                  :string           not null
#  created_at             :datetime         not null
#  application_id         :bigint
#  resource_owner_id      :bigint           not null
#
# Indexes
#
#  index_oauth_access_tokens_on_refresh_token      (refresh_token) UNIQUE
#  index_oauth_access_tokens_on_resource_owner_id  (resource_owner_id)
#  index_oauth_access_tokens_on_token              (token) UNIQUE
#
# Foreign Keys
#
#  fk_rails_...  (application_id => oauth_applications.id)
#  fk_rails_...  (resource_owner_id => users.id)
#
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
