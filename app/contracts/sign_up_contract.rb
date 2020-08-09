# frozen_string_literal: true

class SignUpContract < ApplicationContract
  params do
    required(:email).filled(:stripped_string)
  end

  rule(:email).validate(:email_format)
end
