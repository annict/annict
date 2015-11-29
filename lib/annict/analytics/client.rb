module Annict
  module Analytics
    class Client
      def initialize(request, user)
        @request = request
        @user = user
      end

      def events
        @event ||= Annict::Analytics::Event.new(@request, @user)
      end
    end
  end
end
