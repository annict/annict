# frozen_string_literal: true

module Forms
  class SupporterForm < Forms::ApplicationForm
    attr_writer :auth

    validates :oauth_email, presence: true
    validates :provider_name, presence: true
    validates :provider_uid, presence: true
    validates :provider_token, presence: true
    validates :gumroad_subscriber_id, presence: true
    validates :gumroad_product_id, presence: true
    validates :gumroad_product_name, presence: true
    validates :gumroad_purchase_ids, presence: true
    validates :gumroad_created_at, presence: true

    def oauth_email
      @oauth_email ||= @auth&.dig(:extra, :raw_info, :user, :email)
    end

    def provider_name
      @auth[:provider]
    end

    def provider_uid
      @auth[:uid]
    end

    def provider_token
      @auth[:credentials][:token]
    end

    def subscriber
      @subscriber ||= fetch_subscriber(product_id_jp).presence || fetch_subscriber(product_id)
    end

    def gumroad_subscriber_id
      subscriber&.dig("id")
    end

    def gumroad_product_id
      subscriber&.dig("product_id")
    end

    def gumroad_product_name
      subscriber&.dig("product_name")
    end

    def gumroad_user_id
      subscriber&.dig("user_id")
    end

    def gumroad_user_email
      subscriber&.dig("user_email")
    end

    def gumroad_purchase_ids
      subscriber&.dig("purchase_ids")
    end

    def gumroad_created_at
      subscriber&.dig("created_at")
    end

    def gumroad_cancelled_at
      subscriber&.dig("cancelled_at")
    end

    def gumroad_user_requested_cancellation_at
      subscriber&.dig("user_requested_cancellation_at")
    end

    def gumroad_charge_occurrence_count
      subscriber&.dig("charge_occurrence_count")
    end

    def gumroad_ended_at
      subscriber&.dig("ended_at")
    end

    private

    def product_id
      ENV.fetch("GUMROAD_PRODUCT_ID")
    end

    def product_id_jp
      ENV.fetch("GUMROAD_PRODUCT_ID_JP")
    end

    def fetch_subscriber(product_id)
      client = GumroadClient.new(access_token: ENV.fetch("GUMROAD_ACCESS_TOKEN"))
      json = client.fetch_subscribers(product_id, oauth_email)

      if json["success"] != true
        return nil
      end

      json["subscribers"].filter { |subscriber| subscriber["product_id"] == product_id }.first
    end
  end
end
