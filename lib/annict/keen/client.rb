# frozen_string_literal: true

module Annict
  module Keen
    class Client
      attr_writer :page_category

      ACTIONS = {
        episode_record_create: "episode_record.create",
        status_create: "status.create",
        work_record_create: "work_record.create"
      }.freeze

      def initialize(request, user)
        @request = request
        @user = user
      end

      def publish(action, data = {})
        SendKeenEventJob.perform_later(ACTIONS.fetch(action), base_data.merge(data))
      end

      private

      def base_data
        {
          page_category: @page_category,
          viewer_type: @user ? "user" : "guest",
          user_id: @user&.encoded_id,
          user_weeks: @user&.weeks,
          is_supporter: @user&.supporter? == true,
          time_zone: @user&.time_zone,
          locale: @user&.locale,
          device: browser.device.mobile? ? "mobile" : "pc"
        }
      end

      def browser
        options = {
          accept_language: @request.accept_language
        }
        @browser ||= Browser.new(@request.user_agent, options)
      end
    end
  end
end
