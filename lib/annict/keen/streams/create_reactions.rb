# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CreateReactions < ApplicationStream
        def properties
          base_properties.merge(resource_type: @params[:resource_type], kind: @params[:kind])
        end
      end
    end
  end
end
