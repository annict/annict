# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class StarsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i[create]

      def index
        return render(json: []) unless user_signed_in?

        character_stars = current_user.character_favorites.map { |fav| {starrable_type: "Character", starrable_id: fav.character_id} }
        person_stars = current_user.person_favorites.map { |fav| {starrable_type: "Person", starrable_id: fav.person_id} }
        organization_stars = current_user.organization_favorites.map { |fav| {starrable_type: "Organization", starrable_id: fav.organization_id} }

        render(json: character_stars + person_stars + organization_stars)
      end

      def create
        starrable = params[:starrable_type].constantize.find(params[:starrable_id])
        current_user.favorite(starrable)

        render(json: {}, status: 201)
      end
    end
  end
end
