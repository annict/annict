# frozen_string_literal: true

class SignUpForm < ApplicationForm
  attr_accessor :back
  attr_reader :email

  validates :email, presence: true, email: true
  validate :back_must_be_a_path

  def email=(email)
    @email = email&.strip
  end

  def back_must_be_a_path
    if back && !%r{\A/}.match?(back)
      errors.add(:back, :invalid_back_format)
    end
  end
end
