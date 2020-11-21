# frozen_string_literal: true

module Api
  module Internal
    class ProgramListModalContentsController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      def show
        @program_entities = ProgramListModalContent::ProgramsRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(anime_id: params[:anime_id])
      end
    end
  end
end
