# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CloseTips < ApplicationStream
        def properties
          base_properties.merge(slug: @params[:slug])
        end
      end
    end
  end
end
