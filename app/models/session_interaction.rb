# frozen_string_literal: true

# == Schema Information
#
# Table name: session_interactions
#
#  id         :bigint           not null, primary key
#  email      :citext           not null
#  expires_at :datetime         not null
#  kind       :string           not null
#  token      :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_session_interactions_on_email  (email) UNIQUE
#  index_session_interactions_on_token  (token) UNIQUE
#
class SessionInteraction < ApplicationRecord
  extend Enumerize

  enumerize :kind, in: %i(sign_up sign_in)

  validates :email,
    presence: true,
    uniqueness: { case_sensitive: false },
    email: true
  validates :token, presence: true
  validates :expires_at, presence: true

  def self.start_sign_up!(email:, locale:)
    user = User.only_kept.find_by(email: email)
    return if user

    ActiveRecord::Base.transaction do
      session_interaction = start!(:sign_up, email)
      SessionInteractionMailer.sign_up_interaction(session_interaction, locale).deliver_later
    end
  end

  def self.start_sign_in!(email:, locale:)
    user = User.only_kept.find_by(email: email)
    return unless user

    ActiveRecord::Base.transaction do
      session_interaction = start!(:sign_in, email)
      SessionInteractionMailer.sign_in_interaction(session_interaction, locale).deliver_later
    end
  end

  private_class_method def self.start!(kind, email)
    session_interaction = SessionInteraction.where(email: email).first_or_initialize
    session_interaction.attributes = {
      kind: kind,
      token: SecureRandom.uuid,
      expires_at: Time.zone.now + 2.hours
    }

    session_interaction.save!

    session_interaction
  end

  def expired?
    expires_at.past?
  end
end
