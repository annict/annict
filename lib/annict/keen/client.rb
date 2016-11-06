# frozen_string_literal: true

module Annict
  module Keen
    class Client
      def initialize(request)
        @request = request
      end

      def users
        @users ||= ::Annict::Keen::Events::User.new(@request)
      end

      def tips
        @tips ||= ::Annict::Keen::Events::Tip.new(@request)
      end

      def likes
        @likes ||= ::Annict::Keen::Events::Like.new(@request)
      end
    end
  end
end
