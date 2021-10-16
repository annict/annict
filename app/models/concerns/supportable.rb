# frozen_string_literal: true

module Supportable
  extend ActiveSupport::Concern

  included do
    validates :gumroad_subscriber_id, presence: true
    validates :gumroad_product_id, presence: true
    validates :gumroad_product_name, presence: true
    validates :gumroad_purchase_ids, presence: true
    validates :gumroad_created_at, presence: true

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

    def gumroad_client
      @gumroad_client ||= GumroadClient.new(access_token: ENV.fetch("GUMROAD_ACCESS_TOKEN"))
    end
  end
end
