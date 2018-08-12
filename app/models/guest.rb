# frozen_string_literal: true
# == Schema Information
#
# Table name: guests
#
#  id         :bigint(8)        not null, primary key
#  uuid       :string           not null
#  user_agent :string           default(""), not null
#  remote_ip  :string           default(""), not null
#  time_zone  :string           not null
#  locale     :string           not null
#  aasm_state :string           default("published"), not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_guests_on_uuid  (uuid) UNIQUE
#

class Guest < ApplicationRecord
  include AASM

  has_many :oauth_access_tokens, class_name: "Doorkeeper::AccessToken", dependent: :destroy

  aasm do
    state :published, initial: true
    state :hidden

    event :hide do
      transitions from: :published, to: :hidden
    end
  end

  def find_or_create_access_token_for_official_app!
    @find_or_create_access_token_for_official_app ||= begin
      official_app = Doorkeeper::Application.official
      oauth_access_tokens.
        available.
        where(application: official_app).
        first_or_create!(scopes: official_app.scopes.to_s)
    end
  end
end
