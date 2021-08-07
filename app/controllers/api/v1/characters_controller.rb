# frozen_string_literal: true

module Api
  module V1
    class CharactersController < Api::V1::ApplicationController
      before_action :prepare_params!, only: %i[index]

      def index
        @characters = Character.only_kept
        @characters = Api::V1::CharacterIndexService.new(@characters, @params).result
      end
    end
  end
end
