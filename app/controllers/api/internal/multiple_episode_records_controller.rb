# frozen_string_literal: true

module Api
  module Internal
    class MultipleEpisodeRecordsController < Api::Internal::ApplicationController
      before_action :authenticate_user!

      def create
        episodes = Episode.only_kept.where(id: params[:episode_ids]).order(:sort_number)

        ActiveRecord::Base.transaction do
          episodes.each do |episode|
            EpisodeRecordCreator.new(user: current_user, episode: episode).call
          end
        end

        head 201
      end
    end
  end
end
