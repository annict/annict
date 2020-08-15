# frozen_string_literal: true

class UserEmailContract < ApplicationContract
  params do
    required(:email).filled(:stripped_string)
  end

  rule(:email).validate(:email_format, :email_exists)
end
