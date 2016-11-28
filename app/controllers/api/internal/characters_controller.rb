# frozen_string_literal: true

module Api
  module Internal
    class CharactersController < Api::Internal::ApplicationController
      def index(q: nil)
        @characters = if q.present?
          Character.where("name ILIKE ?", "%#{q}%").published
        else
          Character.none
        end
      end
    end
  end
end
