# frozen_string_literal: true

class RegistrationForm < ApplicationForm
  attr_accessor :token, :terms_and_privacy_policy_agreement
  attr_reader :email, :username

  validates :email,
    presence: true,
    email: true
  validates :terms_and_privacy_policy_agreement,
    presence: true,
    acceptance: true
  validates :username,
    presence: true,
    length: {maximum: 20},
    format: {with: User::USERNAME_FORMAT}

  validate :email_uniqueness
  validate :username_uniqueness

  def email=(email)
    @email = email&.strip
  end

  def username=(username)
    @username = username&.strip
  end

  def email_uniqueness
    if User.find_by(email: email)
      errors.add(:email, :email_uniqueness)
    end
  end

  def username_uniqueness
    if User.find_by(username: username)
      errors.add(:username, :username_uniqueness)
    end
  end
end
