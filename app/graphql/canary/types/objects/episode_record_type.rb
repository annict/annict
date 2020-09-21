# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class EpisodeRecordType < Canary::Types::Objects::Base
        field :rating, Canary::Types::Enums::Rating, null: true, method: :rating_state
        field :advanced_rating, Float, null: true, method: :rating
        field :comments_count, Integer, null: false
        field :episode, Canary::Types::Objects::EpisodeType, null: false

        def episode
          RecordLoader.for(Episode).load(object.episode_id)
        end
      end
    end
  end
end
