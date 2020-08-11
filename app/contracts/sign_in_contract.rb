# frozen_string_literal: true

class SignInContract < ApplicationContract
  params do
    required(:email).filled(:stripped_string)
  end

  rule(:email).validate(:email_format)
end
