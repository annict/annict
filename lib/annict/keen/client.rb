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
    end
  end
end
