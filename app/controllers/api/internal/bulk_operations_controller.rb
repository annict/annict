# frozen_string_literal: true

module Api
  module Internal
    class BulkOperationsController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!, only: %i(show)

      def show
        bulk_operation_entity = BulkOperationRepository.new(graphql_client: graphql_client).execute(job_id: params[:job_id])

        render json: {
          job_id: bulk_operation_entity&.job_id
        }
      end
    end
  end
end
