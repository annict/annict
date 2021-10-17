# frozen_string_literal: true

module Api
  module Internal
    class EpisodeRecordsController < Api::Internal::ApplicationController
      before_action :authenticate_user!, only: %i[create]

      def create
        episode = Episode.only_kept.find(params[:episode_id])
        form = Forms::EpisodeRecordForm.new(user: current_user, episode: episode)
        form.attributes = {
          share_to_twitter: current_user.share_record_to_twitter?
        }

        if form.invalid?
          return render(status: 400, json: {message: form.errors.full_messages.first})
        end

        creator = Creators::EpisodeRecordCreator.new(user: current_user, form: form).call

        render(status: 201, json: {record_id: creator.record.id})
      end
    end
  end
end
