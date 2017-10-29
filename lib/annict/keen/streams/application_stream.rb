# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class ApplicationStream
        attr_reader :request

        def initialize(request, params)
          @request = request
          @params = params
        end

        def browser
          options = {
            accept_language: @request.accept_language
          }
          @browser ||= Browser.new(@request.user_agent, options)
        end

        def locale
          user&.locale.presence || @params[:locale]
        end

        def time_zone
          user&.time_zone.presence || @params[:time_zone]
        end

        def page_category
          @params[:page_category]
        end

        def user
          @params[:user]
        end

        def via
          @params[:via]
        end

        def timestamp
          Time.zone.now.to_s
        end

        def base_properties
          {
            device: browser.device.mobile? ? "mobile" : "pc",
            keen: { timestamp: timestamp },
            locale: locale,
            page_category: page_category,
            time_zone: time_zone,
            user_agent: request.user_agent,
            user_id: user&.encoded_id,
            uuid: request.cookies["ann_client_uuid"],
            via: via
          }
        end
      end
    end
  end
end
