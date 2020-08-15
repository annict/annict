# frozen_string_literal: true

class RegistrationContract < ApplicationContract
  params do
    required(:email).filled(:string)
    required(:username).filled(:stripped_string)
    required(:token).filled(:string)
    required(:terms_and_privacy_policy_agreement).filled(:bool)
  end

  rule(:email).validate(:email_format, :email_exists)
  rule(:username).validate(:username_format)
  rule(:terms_and_privacy_policy_agreement) do
    unless value
      key.failure(:acceptance)
    end
  end
end
