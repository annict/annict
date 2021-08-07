# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class EpisodeRecordType < Canary::Types::Objects::Base
        field :rating, Canary::Types::Enums::Rating, null: true
        field :advanced_rating, Float, null: true, method: :rating
        field :comments_count, Int, null: false
        field :episode, Canary::Types::Objects::EpisodeType, null: false

        def rating
          if object.rating
            object.rating_to_rating_state.to_s.presence
          else
            object.rating_state
          end
        end

        def episode
          RecordLoader.for(Episode).load(object.episode_id)
        end
      end
    end
  end
end
