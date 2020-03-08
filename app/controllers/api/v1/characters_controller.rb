# frozen_string_literal: true

module API
  module V1
    class CharactersController < API::V1::ApplicationController
      before_action :prepare_params!, only: %i(index)

      def index
        @characters = Character.without_deleted
        @characters = API::V1::CharacterIndexService.new(@characters, @params).result
      end
    end
  end
end
