# frozen_string_literal: true

class SignInContract < ApplicationContract
  params do
    required(:email).filled(:stripped_string)
    required(:back).maybe(:string)
  end

  rule(:email).validate(:email_format)
  rule(:back).validate(:back_format)
end
