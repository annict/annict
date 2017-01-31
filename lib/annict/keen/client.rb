# frozen_string_literal: true

module Annict
  module Keen
    class Client
      def initialize(request)
        @request = request
      end

      def follows
        @follows ||= ::Annict::Keen::Events::Follow.new(@request)
      end

      def likes
        @likes ||= ::Annict::Keen::Events::Like.new(@request)
      end

      def multiple_records
        @multiple_records ||= ::Annict::Keen::Events::MultipleRecord.new(@request)
      end

      def statuses
        @statuses ||= ::Annict::Keen::Events::Status.new(@request)
      end

      def tips
        @tips ||= ::Annict::Keen::Events::Tip.new(@request)
      end

      def users
        @users ||= ::Annict::Keen::Events::User.new(@request)
      end
    end
  end
end
