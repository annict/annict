# frozen_string_literal: true

module API
  module Internal
    class CharactersController < API::Internal::ApplicationController
      def index
        @characters = if params[:q].present?
          Character.where("name ILIKE ?", "%#{params[:q]}%").without_deleted
        else
          Character.none
        end
      end
    end
  end
end
