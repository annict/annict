# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CreateStatuses < ApplicationStream
        def properties
          base_properties.merge(
            work_id: @params[:work_id],
            kind: @params[:kind],
            is_first_status: @params[:is_first_status],
            oauth_application_id: @params[:oauth_application_id]
          )
        end
      end
    end
  end
end
