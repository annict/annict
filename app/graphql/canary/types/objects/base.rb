# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class Base < GraphQL::Schema::Object
        include Imgix::Rails::UrlHelper
        include ImageHelper

        def annict_id
          object.id
        end
      end
    end
  end
end
