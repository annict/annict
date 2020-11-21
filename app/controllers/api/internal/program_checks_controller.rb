# frozen_string_literal: true

module Api
  module Internal
    class ProgramChecksController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      def create
        _, err = CheckProgramRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(program_id: params[:program_id])

        if err
          return render(status: 400, json: { message: err.message })
        end

        head 201
      end
    end
  end
end
