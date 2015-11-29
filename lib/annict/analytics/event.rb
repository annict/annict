module Annict
  module Analytics
    class Event
      include HTTParty

      base_uri "https://ssl.google-analytics.com"

      def initialize(request)
        @request = request
      end

      # イベントを送る
      # https://developers.google.com/analytics/devguides/collection/protocol/v1/devguide#event
      # ec: Event Category. Required.
      # ea: Event Action. Required.
      # el: Event label.
      # ev: Event value.
      def create(ec, ea, el: "", ev: "")
        headers = { "User-Agent" => @request.user_agent || "" }
        body = {
          v: 1,
          tid: ENV.fetch("GA_TRACKING_ID"),
          cid: SecureRandom.uuid, # TODO: 雑にcidを生成しちゃってるけどこれで良いのか?
          t: "event",
          ec: ec,
          ea: ea
        }
        body[:el] = el if el.present?
        body[:ev] = ev if ev.present?

        self.class.delay.post("/collect", headers: headers, body: body)
      end
    end
  end
end
