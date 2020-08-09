# frozen_string_literal: true

class RegistrationForm < ApplicationForm
  attr_accessor :email, :token, :username, :terms_and_privacy_policy_agreement
end
