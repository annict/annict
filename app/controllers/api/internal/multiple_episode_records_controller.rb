# frozen_string_literal: true

module Api
  module Internal
    class MultipleEpisodeRecordsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        episodes = Episode.only_kept.where(id: params[:episode_ids]).order(:sort_number)

        ActiveRecord::Base.transaction do
          episodes.each do |episode|
            form = Forms::EpisodeRecordForm.new(episode: episode, share_to_twitter: false)
            form.user = current_user
            Creators::EpisodeRecordCreator.new(user: current_user, form: form).call
          end
        end

        head 201
      end
    end
  end
end
