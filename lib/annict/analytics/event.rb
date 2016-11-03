# frozen_string_literal: true

module Annict
  module Analytics
    class Event
      include HTTParty

      base_uri "https://ssl.google-analytics.com"

      def initialize(request, user)
        @request = request
        @user = user
      end

      # イベントを送る
      # https://developers.google.com/analytics/devguides/collection/protocol/v1/devguide#event
      # ec: Event Category. Required.
      # ea: Event Action. Required.
      # el: Event label.
      # ev: Event value.
      # ds: Data source.
      def create(ec, ea, el: "", ev: "", ds: "web")
        body = {
          v: 1,
          tid: ENV.fetch("GA_TRACKING_ID"),
          cid: @request.cookies["ann_ga_cid"],
          t: "event",
          ec: ec,
          ea: ea,
          uip: @request.ip,
          ua: @request.user_agent,
          ds: ds
        }
        body[:uid] = @user.encoded_id if @user.present?
        body[:el] = el if el.present?
        body[:ev] = ev if ev.present?

        self.class.delay(priority: 10).post("/collect", body: body)
      end
    end
  end
end
