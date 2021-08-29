# frozen_string_literal: true

module Api
  module V1
    module Me
      class ReviewsUpdateParams
        include ActiveParameter

        param :id
        param :title
        param :body
        param :rating_animation_state
        param :rating_music_state
        param :rating_story_state
        param :rating_character_state
        param :rating_overall_state
        param :share_twitter, default: "false"
        param :share_facebook, default: "false"

        validates :id,
          presence: true,
          numericality: {
            only_integer: true,
            greater_than_or_equal_to: 1
          }
        validates :body,
          presence: true
        validates :share_twitter,
          allow_blank: true,
          filter_boolean_params: true
        validates :share_facebook,
          allow_blank: true,
          filter_boolean_params: true
      end
    end
  end
end
