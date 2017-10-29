# frozen_string_literal: true

module Annict
  module Keen
    module Streams
      class CreateFollows < ApplicationStream
        def properties
          followed_user = @params[:followed_user]
          base_properties.merge(followed_user_id: followed_user.encoded_id)
        end
      end
    end
  end
end
