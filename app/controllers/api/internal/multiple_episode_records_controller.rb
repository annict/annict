# frozen_string_literal: true

module Api
  module Internal
    class MultipleEpisodeRecordsController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!, only: %i(create)

      def create
        form = MultipleEpisodeRecordForm.new(episode_ids: params[:episode_ids])

        bulk_operation_entity, err = BulkCreateEpisodeRecordsRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(form: form)

        if err
          return render(status: 400, json: { message: err.message })
        end

        render json: {
          job_id: bulk_operation_entity.job_id
        }
      end
    end
  end
end
