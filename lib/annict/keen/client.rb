# frozen_string_literal: true

module Annict
  module Keen
    class Client
      attr_writer :app
      attr_writer :page_category
      attr_writer :user

      def initialize(request, user)
        @request = request
        @user = user
      end

      def follows
        @follows ||= ::Annict::Keen::Events::Follow.new(@request, @user, params)
      end

      def likes
        @likes ||= ::Annict::Keen::Events::Like.new(@request, @user, params)
      end

      def multiple_records
        @multiple_records ||= ::Annict::Keen::Events::MultipleRecord.new(@request, @user, params)
      end

      def statuses
        @statuses ||= ::Annict::Keen::Events::Status.new(@request, @user, params)
      end

      def tips
        @tips ||= ::Annict::Keen::Events::Tip.new(@request, @user, params)
      end

      def users
        @users ||= ::Annict::Keen::Events::User.new(@request, @user, params)
      end

      private

      def params
        {
          app: @app,
          page_category: @page_category
        }
      end
    end
  end
end
