# frozen_string_literal: true

# == Schema Information
#
# Table name: email_confirmations
#
#  id         :bigint           not null, primary key
#  back       :string
#  email      :citext           not null
#  event      :string           not null
#  expires_at :datetime         not null
#  token      :string           not null
#  created_at :datetime         not null
#  updated_at :datetime         not null
#
# Indexes
#
#  index_email_confirmations_on_token  (token) UNIQUE
#
class EmailConfirmation < ApplicationRecord
  extend Enumerize

  enumerize :event, in: %i(sign_up sign_in)

  validates :email, presence: true, email: true
  validates :token, presence: true
  validates :expires_at, presence: true

  def confirm_to_sign_up!
    user = User.only_kept.find_by(email: email)
    return if user

    ActiveRecord::Base.transaction do
      confirmation = create_confirmation!(event: :sign_up)
      EmailConfirmationMailer.sign_up_confirmation(confirmation.id, I18n.locale.to_s).deliver_later
    end
  end

  def confirm_to_sign_in!
    user = User.only_kept.find_by(email: email)
    return unless user

    ActiveRecord::Base.transaction do
      confirmation = create_confirmation!(event: :sign_in)
      EmailConfirmationMailer.sign_in_confirmation(confirmation.id, I18n.locale.to_s).deliver_later
    end
  end

  def expired?
    expires_at.past?
  end

  private

  def create_confirmation!(event:)
    self.class.create!(
      email: email,
      event: event,
      token: SecureRandom.uuid,
      back: back,
      expires_at: Time.zone.now + 2.hours
    )
  end
end
