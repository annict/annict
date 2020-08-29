# frozen_string_literal: true

module Api
  module Internal
    class EpisodeRecordsController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!, only: %i(create)

      def create
        episode = Episode.only_kept.find(params[:episode_id])

        episode_record, err = CreateEpisodeRecordRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(
          episode: episode,
          params: {
            rating_state: episode_record_params[:rating_state],
            body: episode_record_params[:body],
            share_to_twitter: episode_record_params[:shared_twitter]
          }
        )

        if err
          return render(status: 400, json: { message: err.message })
        end

        head 201
      end

      private

      def episode_record_params
        params.require(:episode_record).permit(:body, :shared_twitter, :rating_state)
      end
    end
  end
end
