module Annict
  module Analytics
    class Client
      def initialize(request)
        @request = request
      end

      def events
        @event ||= Annict::Analytics::Event.new(@request)
      end
    end
  end
end
