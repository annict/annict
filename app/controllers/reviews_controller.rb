# frozen_string_literal: true

class ReviewsController < ApplicationController
  permits :title, :body, :rating_animation_state, :rating_music_state, :rating_story_state,
    :rating_character_state, :rating_overall_state

  impressionist actions: %i(show)

  def show
    user = User.find_by!(username: params[:username])
    review = user.reviews.published.find(params[:id])
    redirect_to record_path(user.username, review.record_id), status: 301
  end
end
