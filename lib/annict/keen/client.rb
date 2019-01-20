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
        SendKeenEventJob.perform_later(ACTIONS.fetch(action), build_data(data))
      end

      private

      def build_data(data)
        data = data.merge(
          page_category: @page_category.presence || "unknown",
          viewer: {
            type: @user.present? ? "user" : "guest"
          }
        )
        data[:user] = { id: @user.id, username: @user.username } if @user
        data
      end
    end
  end
end
