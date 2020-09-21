# frozen_string_literal: true

module Canary
  module Types
    module Objects
      class AnimeRecordType < Canary::Types::Objects::Base
        field :rating_overall, Canary::Types::Enums::Rating, null: true
        field :rating_animation, Canary::Types::Enums::Rating, null: true
        field :rating_music, Canary::Types::Enums::Rating, null: true
        field :rating_story, Canary::Types::Enums::Rating, null: true
        field :rating_character, Canary::Types::Enums::Rating, null: true
      end
    end
  end
end
