# frozen_string_literal: true

class SignUpForm < ApplicationForm
  attr_accessor :email, :back

  validates :email, presence: true, email: true
  validate :back_must_be_a_path

  def back_must_be_a_path
    if back && !%r{\A/}.match?(back)
      errors.add(:back, :invalid_back_format)
    end
  end
end
