# frozen_string_literal: true

module Annict
  module Keen
    module Events
      class Application
        attr_reader :request

        def initialize(request)
          @request = request
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
end
