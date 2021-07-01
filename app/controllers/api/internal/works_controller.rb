# frozen_string_literal: true

module Api
  module Internal
    class WorksController < Api::Internal::ApplicationController
      def index
        q = params[:q]
        @works = if q
          Anime.where("title ILIKE ?", "%#{q}%").only_kept
        else
          Anime.none
        end
      end
    end
  end
end
