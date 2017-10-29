# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CreateItems < ApplicationStream
        def properties
          base_properties.merge(resource_type: @params[:resource_type])
        end
      end
    end
  end
end
