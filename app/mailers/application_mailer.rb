# frozen_string_literal: true
# typed: false

class ApplicationMailer < ActionMailer::Base
  default from: "from@example.com"
  layout "mailer"
end
