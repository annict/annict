# frozen_string_literal: true

module Annict
  module Keen
    class Client
      def initialize(request)
        @request = request
      end

      def publish(stream_name, params)
        @stream_name = stream_name
        @params = params

        SendEventsToKeenJob.perform_later(stream_name, stream.properties)
      end

      private

      def stream
        stream_class = Object.const_get("Annict::Keen::Streams::#{@stream_name.camelcase}")
        stream_class.new(@request, @params)
      end
    end
  end
end
