# frozen_string_literal: true

module AnalyticsFilter
  extend ActiveSupport::Concern

  included do
    before_action :store_ann_ga_cid

    def ga_client
      @ga_client ||= Annict::Analytics::Client.new(request, current_user)
    end

    private

    def store_ann_ga_cid
      return if cookies[:ann_ga_cid].present?

      domain = Rails.env.development? ? "localhost" : ".annict.com"
      cookies[:ann_ga_cid] = {
        value: SecureRandom.uuid,
        expires: 2.year.from_now,
        domain: domain
      }
    end
  end
end
