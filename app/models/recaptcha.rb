# frozen_string_literal: true

class Recaptcha
  SITE_KEY = Rails.env.production? ? ENV.fetch("RECAPTCHA_SITE_KEY") : ENV["RECAPTCHA_SITE_KEY"]
  SECRET_KEY = Rails.env.production? ? ENV.fetch("RECAPTCHA_SECRET_KEY") : ENV["RECAPTCHA_SECRET_KEY"]
  MINIMUM_SCORE = 0.5

  attr_reader :action

  def self.enabled?
    [SITE_KEY, SECRET_KEY].all?(&:present?)
  end

  def initialize(action:)
    @action = action
  end

  def enabled?
    self.class.enabled?
  end

  def verify?(token)
    unless enabled?
      return true
    end

    uri = URI.parse("https://www.google.com/recaptcha/api/siteverify?secret=#{SECRET_KEY}&response=#{token}")
    response = Net::HTTP.get_response(uri)
    json = JSON.parse(response.body)

    json["success"] && json["score"] > MINIMUM_SCORE && json["action"] == action
  end

  def id
    @id ||= "recaptcha_token_#{SecureRandom.hex(10)}"
  end

  def site_key
    SITE_KEY
  end
end
