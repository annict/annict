# frozen_string_literal: true

class UserEmailForm < ApplicationForm
  attr_accessor :email

  def persisted?
    true
  end
end
