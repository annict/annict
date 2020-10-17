# frozen_string_literal: true

module Api
  module Internal
    class EpisodeRecordsController < Api::Internal::ApplicationController
      include V4::GraphqlRunnable

      before_action :authenticate_user!, only: %i(create)

      def create
        form = EpisodeRecordForm.new(episode_id: params[:episode_id], share_to_twitter: current_user.share_record_to_twitter?)

        episode_record, err = CreateEpisodeRecordRepository.new(
          graphql_client: graphql_client(viewer: current_user)
        ).execute(form: form)

        if err
          return render(status: 400, json: { message: err.message })
        end

        head 201
      end
    end
  end
end
