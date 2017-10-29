# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CreateReviews < ApplicationStream
        def properties
          base_properties.merge(
            work_id: @params[:work_id],
            oauth_application_id: @params[:oauth_application_id]
          )
        end
      end
    end
  end
end
