# frozen_string_literal: true

class UserEmailForm < ApplicationForm
  attr_reader :email

  validates :email,
    presence: true,
    email: true
  validate :email_uniqueness

  def email=(email)
    @email = email&.strip
  end

  def email_uniqueness
    if User.find_by(email: email)
      errors.add(:email, :email_uniqueness)
    end
  end

  def persisted?
    true
  end
end
