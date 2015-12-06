module Annict
  module Keen
    module Collections
      class ApplicationCollection
        attr_reader :request

        def initialize(request)
          @request = request
        end

        def browser
          options = {
            user_agent: @request.user_agent,
            accept_language: @request.accept_language
          }

          @browser ||= Browser.new(options)
        end
      end
    end
  end
end
