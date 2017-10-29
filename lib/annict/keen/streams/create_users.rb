# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CreateUsers < ApplicationStream
        def properties
          base_properties.merge(via_oauth: @params[:via_oauth])
        end
      end
    end
  end
end
