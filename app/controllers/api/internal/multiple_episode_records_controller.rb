# frozen_string_literal: true

module Api
  module Internal
    class MultipleEpisodeRecordsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        episodes = Episode.only_kept.where(id: params[:episode_ids]).order(:sort_number)

        ActiveRecord::Base.transaction do
          episodes.each do |episode|
            form = Forms::RecordForm.new(episode_id: episode.id, instant: true, skip_to_share: true)
            Creators::RecordCreator.new(user: current_user, form: form).call
          end
        end

        head 201
      end
    end
  end
end
