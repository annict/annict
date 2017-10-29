# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CreateMultipleRecords < ApplicationStream
        def properties
          base_properties
        end
      end
    end
  end
end
