# frozen_string_literal: true

class SignInForm < ApplicationForm
  attr_accessor :back
  attr_reader :email

  validates :back, format: {with: %r{\A/}, allow_blank: true}
  validates :email, presence: true, email: true

  def email=(email)
    @email = email&.strip
  end
end
