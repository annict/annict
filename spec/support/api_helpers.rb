# frozen_string_literal: true

module APIHelpers
  def api(path, data = {})
    return path if data.blank?

    params = data.map { |key, val| "#{key}=#{val}" }.join("&")
    "#{path}?#{params}"
  end

  def json
    JSON.parse(response.body)
  end
end

RSpec.configure do |config|
  config.before :each, type: :request do
    host! "api.annict.test:3000"
  end

  config.include APIHelpers, type: :request
end
