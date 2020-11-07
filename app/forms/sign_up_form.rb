# frozen_string_literal: true

class SignUpForm < ApplicationForm
  attr_accessor :back
  attr_reader :email

  validates :back, format: { with: %r{\A/} }
  validates :email, presence: true, email: true

  def email=(email)
    @email = email&.strip
  end
end
