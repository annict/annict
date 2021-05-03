# frozen_string_literal: true

class GumroadSubscribersSyncService
  API_ENDPOINT = "https://api.gumroad.com"

  def self.execute
    new.execute
  end

  def execute
    [
      ENV.fetch("GUMROAD_PRODUCT_ID"),
      ENV.fetch("GUMROAD_PRODUCT_ID_JP")
    ].each do |product_id|
      path = "/v2/products/#{product_id}/subscribers"
      fetch_and_save(path)
    end
  end

  private

  def fetch_and_save(path)
    url = "#{API_ENDPOINT}#{path}"
    options = {
      headers: {
        "Authorization" => "Bearer #{ENV.fetch("GUMROAD_ACCESS_TOKEN")}"
      }
    }
    response = HTTParty.get(url, options)
    json = JSON.parse(response.body)

    if json["success"] != true
      message = "@everyone Failed to fetch from `#{url}`"
      message_to_discord(message)
      return
    end

    json["subscribers"].each do |subscriber|
      gs = GumroadSubscriber.where(gumroad_id: subscriber["id"]).first_or_initialize

      gs.gumroad_product_id = subscriber["product_id"]
      gs.gumroad_product_name = subscriber["product_name"]
      gs.gumroad_user_id = subscriber["user_id"]
      gs.gumroad_user_email = subscriber["user_email"]
      gs.gumroad_purchase_ids = subscriber["purchase_ids"]
      gs.gumroad_created_at = subscriber["created_at"]
      gs.gumroad_cancelled_at = subscriber["cancelled_at"]
      gs.gumroad_user_requested_cancellation_at = subscriber["user_requested_cancellation_at"]
      gs.gumroad_charge_occurrence_count = subscriber["charge_occurrence_count"]
      gs.gumroad_ended_at = subscriber["ended_at"]

      next if gs.save

      message = <<~ERROR
        Validation Failed!
        messages: #{gs.errors.full_messages}
        subscriber: #{subscriber}
      ERROR
      message_to_discord(message)
    end

    return if json["next_page_url"].blank?

    fetch_and_save(json["next_page_url"])
  end

  def message_to_discord(message)
    options = {
      url: ENV.fetch("ANNICT_DISCORD_WEBHOOK_URL_FOR_GUMROAD")
    }
    Discord::Notifier.message(message, options)
  end
end
