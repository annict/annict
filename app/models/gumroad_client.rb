# frozen_string_literal: true

class GumroadClient
  API_ENDPOINT = "https://api.gumroad.com"
  PRODUCT_ID = ENV.fetch("GUMROAD_PRODUCT_ID")
  PRODUCT_ID_JP = ENV.fetch("GUMROAD_PRODUCT_ID_JP")

  def initialize(access_token:)
    @access_token = access_token
  end

  def fetch_subscriber_by_email(email)
    fetch_subscriber(PRODUCT_ID_JP, email).presence || fetch_subscriber(PRODUCT_ID, email)
  end

  def fetch_subscriber_by_subscriber_id(subscriber_id)
    response = HTTParty.get("#{API_ENDPOINT}/v2/subscribers/#{subscriber_id}", headers: headers)
    json = JSON.parse(response.body)

    if json["success"] != true
      return nil
    end

    json["subscriber"]
  end

  private

  def headers
    {
      "Authorization" => "Bearer #{@access_token}"
    }
  end

  def fetch_subscriber(product_id, email)
    response = HTTParty.get("#{API_ENDPOINT}/v2/products/#{product_id}/subscribers", {
      headers: headers,
      query: {
        email: email
      }
    })
    json = JSON.parse(response.body)

    if json["success"] != true
      return nil
    end

    json["subscribers"].find { |subscriber| subscriber["product_id"] == product_id }
  end
end
