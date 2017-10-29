# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CreateRecords < ApplicationStream
        def properties
          base_properties.merge(
            work_id: @params[:work_id],
            episode_id: @params[:episode_id],
            has_comment: @params[:has_comment],
            shared_twitter: @params[:shared_twitter],
            shared_facebook: @params[:shared_facebook],
            is_first_record: @params[:is_first_record],
            oauth_application_id: @params[:oauth_application_id]
          )
        end
      end
    end
  end
end
