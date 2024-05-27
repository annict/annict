# typed: false
# frozen_string_literal: true

module Api
  module Internal
    class CharactersController < Api::Internal::ApplicationController
      def index
        @characters = if params[:q].present?
          Character.where("name ILIKE ?", "%#{params[:q]}%").only_kept
        else
          Character.none
        end
      end
    end
  end
end
