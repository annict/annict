# typed: false
# frozen_string_literal: true

module Api
  module V1
    module Me
      class ReviewsCreateParams
        include ActiveParameter

        param :work_id
        param :title
        param :body
        param :rating_animation_state
        param :rating_music_state
        param :rating_story_state
        param :rating_character_state
        param :rating_overall_state
        param :share_twitter, default: "false"
        param :share_facebook, default: "false"

        validates :work_id,
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
        WorkRecord::STATES.each do |state|
          validates state,
            presence: true,
            format: {
              with: /\A(bad|average|good|great)\z/,
              message: "の値が不正です。"
            },
            allow_blank: true
        end
      end
    end
  end
end
