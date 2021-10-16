# frozen_string_literal: true

class GumroadClient
  API_ENDPOINT = "https://api.gumroad.com"

  def initialize(access_token:)
    @access_token = access_token
  end

  def fetch_subscribers(product_id, email)
    response = HTTParty.get("#{API_ENDPOINT}/v2/products/#{product_id}/subscribers", {
      headers: headers,
      query: {
        email: email
      }
    })

    JSON.parse(response.body)
  end

  private

  def headers
    {
      "Authorization" => "Bearer #{@access_token}"
    }
  end
end
