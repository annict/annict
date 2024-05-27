# typed: false
# frozen_string_literal: true

module Annict
  module Analytics
    class Event
      include HTTParty
      include GaHelper

      base_uri "https://ssl.google-analytics.com"

      def initialize(request, user, params)
        @request = request
        @user = user
        @params = params
      end

      # イベントを送る
      # https://developers.google.com/analytics/devguides/collection/protocol/v1/devguide#event
      # ec: Event Category. Required.
      # ea: Event Action. Required.
      # el: Event label. Recommended.
      # ev: Event value.
      # ds: Data source.
      def create(ec, ea, el: "", ev: "", ds: :web)
        body = {
          v: 1,
          tid: ga_tracking_id(@request),
          cid: @request.cookies["ann_client_uuid"],
          t: "event",
          ec: ec.to_s,
          ea: ea.to_s,
          uip: @request.ip,
          ua: @request.user_agent,
          ds: ds.to_s
        }
        body[:uid] = @user.encoded_id if @user.present?
        body[:el] = el.to_s if el.present?
        body[:ev] = ev.to_s if ev.present?
        body[:cd1] = @user.present? ? "user" : "guest"
        body[:cd2] = @params[:page_category].to_s if @params[:page_category].present?

        SendAnalyticsEventJob.perform_later(body)
      end
    end
  end
end
